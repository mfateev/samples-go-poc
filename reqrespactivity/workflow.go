package reqrespactivity

import (
	"context"
	"strings"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
	"go.temporal.io/sdk/temporal"
)

// UppercaseWorkflow is a workflow that accepts requests to uppercase strings
// via signals and provides responses via a callback response activity.
//
// The "request" signal accepts a Request.
//
//temporal:workflow
func UppercaseWorkflow(ctx context.Context) error {
	// Create and run the uppercaser. We choose to use a separate struct for this
	// to make state management easier.
	u, err := newUppercaser(ctx)
	if err == nil {
		err = u.run()
	}
	return err
}

// UppercaseActivity uppercases the given string.
func UppercaseActivity(ctx context.Context, input string) (string, error) {
	return strings.ToUpper(input), nil
}

// Request is a request to uppercase a string, passed as a signal argument to
// UppercaseWorkflow.
type Request struct {
	// ID of the request, also set on the response.
	ID string `json:"id"`
	// String to be uppercased.
	Input string `json:"input"`
	// Required information on which activity to send response to.
	ResponseActivity  string `json:"response_activity"`
	ResponseTaskQueue string `json:"response_task_queue"`
}

// Response is a response to a Request. This is the parameter of the response
// activity.
type Response struct {
	ID     string `json:"id"`
	Output string `json:"output"`
	Error  string `json:"error"`
}

type uppercaser struct {
	ctx                         context.Context
	requestCh                   <-chan Request
	requestsBeforeContinueAsNew int
	responseActivityOptions     workflow.ActivityOptions
}

func newUppercaser(ctx context.Context) (*uppercaser, error) {
	u := &uppercaser{
		// For the main context, we're only going to allow 1 retry and only a 5
		// second schedule-to-close timeout. WARNING: The timeout and retry affect
		// how long this workflow stays open and may prevent it from performing its
		// continue-as-new until timeout occurs and/or retries are finished.
		ctx: workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			ScheduleToCloseTimeout: 5 * time.Second,
			RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 2},
		}),

		// Use a signal to obtain requests. Since this cannot be closed, the
		// workflow counts on the ability to have some "idle" period between
		// handling requests/responses where it can continue-as-new.
		requestCh: workflow.GetSignalChannel[Request](ctx, "request"),

		// We'll allow 500 requests before we continue-as-new the workflow. This is
		// required because the history will grow very large otherwise for an
		// interminable workflow fielding signal requests and executing activities.
		requestsBeforeContinueAsNew: 500,

		// We're gonna just support 10 second timeout and 3 retries on response
		// callback activities. WARNING: The timeout and retry affect how long this
		// workflow stays open and may prevent it from performing its
		// continue-as-new until timeout occurs and/or retries are finished.
		responseActivityOptions: workflow.ActivityOptions{
			// We use schedule-to-start/close because if the requester side is not
			// present, this may hang otherwise with just start-to-close.
			ScheduleToStartTimeout: 5 * time.Second,
			ScheduleToCloseTimeout: 10 * time.Second,
			RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 4},
		},
	}
	return u, nil
}

func (u *uppercaser) run() error {
	// Each goroutine owns one request. The main loop owns bookkeeping and waits
	// for all accepted work before draining signals and continuing as new.
	completed := make(chan struct{})
	count, pending := 0, 0
	start := func(request Request) {
		count++
		pending++
		go func() {
			u.process(request)
			select {
			case completed <- struct{}{}:
			case <-u.ctx.Done():
			}
		}()
	}
	for {
		if err := u.ctx.Err(); err != nil {
			return err
		}
		if count >= u.requestsBeforeContinueAsNew && pending == 0 {
			// Signals admitted in the current task are already in this native buffer.
			select {
			case request := <-u.requestCh:
				start(request)
				continue
			default:
				return workflow.NewContinueAsNewError(u.ctx, UppercaseWorkflow)
			}
		}
		select {
		case <-u.ctx.Done():
			return u.ctx.Err()
		case request := <-u.requestCh:
			start(request)
		case <-completed:
			pending--
		}
	}
}

func (u *uppercaser) process(request Request) {
	response := &Response{ID: request.ID}
	if err := workflow.ExecuteActivity(u.ctx, UppercaseActivity, request.Input).Get(u.ctx, &response.Output); err != nil {
		response.Error = err.Error()
	}
	options := u.responseActivityOptions
	options.TaskQueue = request.ResponseTaskQueue
	ctx := workflow.WithActivityOptions(u.ctx, options)
	if err := workflow.ExecuteActivity(ctx, request.ResponseActivity, response).Get(ctx, nil); err != nil {
		workflow.GetLogger(ctx).Warn("Failure sending response activity", "error", err)
	}
}

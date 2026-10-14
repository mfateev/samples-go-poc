package reqrespquery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
	"go.temporal.io/sdk/temporal"
)

// UppercaseWorkflow is a workflow that accepts requests to uppercase strings
// via signals and provides responses via query.
//
// The "request" signal accepts a Request.
//
// The "response" query accepts a request ID.
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
}

// Response is a response to a Request. This can be returned in a map from a
// query or as the parameter of a callback to a response activity.
type Response struct {
	ID     string `json:"id"`
	Output string `json:"output"`
	Error  string `json:"error"`
}

type uppercaser struct {
	ctx                         context.Context
	requestCh                   <-chan Request
	requestsBeforeContinueAsNew int
	// This maintains responses for the lifetime of the workflow, which should not
	// be too large before continue-as-new.
	responses map[string]*Response
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
		responses:                   make(map[string]*Response, 500),
	}
	// Set query handler
	if err := workflow.SetQueryHandler(ctx, "response", u.queryResponse); err != nil {
		return nil, fmt.Errorf("failed setting query handler: %w", err)
	}
	return u, nil
}

func (u *uppercaser) run() error {
	// Each goroutine owns one request. The main loop owns bookkeeping and waits
	// for all accepted work before draining signals and continuing as new.
	completed := make(chan *Response)
	count, pending := 0, 0
	start := func(request Request) {
		count++
		pending++
		go func() {
			response := u.process(request)
			select {
			case completed <- response:
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
		case response := <-completed:
			pending--
			u.responses[response.ID] = response
		}
	}
}

func (u *uppercaser) queryResponse(id string) (*Response, error) {
	// We intentionally do not remove the response from the map for the
	// following reasons:
	// * It is not acceptable for a query to have side effects
	// * The query may be called multiple times by the requester even when found
	// * A workflow writer should treat their callers as external entities and
	//   should not assume how the query will be used
	return u.responses[id], nil
}

func (u *uppercaser) process(request Request) *Response {
	response := &Response{ID: request.ID}
	if err := workflow.ExecuteActivity(u.ctx, UppercaseActivity, request.Input).Get(u.ctx, &response.Output); err != nil {
		response.Error = err.Error()
	}
	return response
}

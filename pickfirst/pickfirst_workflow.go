package pickfirst

import (
	"context"
	"fmt"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
	"go.temporal.io/sdk/activity"
)

/**
 * This sample workflow execute activities in parallel branches, pick the result of the branch that completes first,
 * and then cancels other activities that are not finished yet.
 */

// SamplePickFirstWorkflow workflow definition
//
//temporal:workflow
func SamplePickFirstWorkflow(ctx context.Context) (string, error) {
	callCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	callCtx = workflow.WithActivityOptions(callCtx, workflow.ActivityOptions{StartToCloseTimeout: 2 * time.Minute, HeartbeatTimeout: 10 * time.Second, WaitForCancellation: true})
	futures := []workflow.Future{workflow.ExecuteActivity(callCtx, SampleActivity, 0, 10*time.Second), workflow.ExecuteActivity(callCtx, SampleActivity, 1, time.Second)}
	var winner workflow.Future
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-futures[0].ToChannel():
		winner = futures[0]
	case <-futures[1].ToChannel():
		winner = futures[1]
	}
	var response string
	err := winner.Get(ctx, &response)
	cancel()
	// Wait for cancellation acknowledgements before completing the workflow.
	for _, f := range futures {
		_ = f.Get(ctx, nil)
	}
	return response, err
}

func SampleActivity(ctx context.Context, currentBranchID int, totalDuration time.Duration) (string, error) {

	logger := activity.GetLogger(ctx)
	elapsedDuration := time.Nanosecond
	for elapsedDuration < totalDuration {
		time.Sleep(time.Second)
		elapsedDuration += time.Second

		// record heartbeat every second to check if we are been cancelled
		activity.RecordHeartbeat(ctx, "status-report-to-workflow")

		select {
		case <-ctx.Done():
			// We have been cancelled.
			msg := fmt.Sprintf("Branch %d is cancelled.", currentBranchID)
			logger.Info(msg)
			return msg, ctx.Err()
		default:
			// We are not cancelled yet.
		}

		// Do some custom work
		// ...
	}

	msg := fmt.Sprintf("Branch %d done in %s.", currentBranchID, totalDuration)
	return msg, nil
}

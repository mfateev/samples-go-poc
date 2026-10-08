// The upstream sleep-for-days workflow, adapted to use native Go channels and
// select inside an isolate.
package sleepfordays

import (
	"context"
	"errors"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
)

//go:isolate
func SleepForDays(ctx context.Context) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Second})
	signals := workflow.GetSignalChannel(ctx, "complete")
	for {
		// The upstream sample schedules the email without awaiting its future.
		// Completion or failure of the email does not control this workflow.
		_ = workflow.ExecuteActivity(ctx, SendEmail, "Sleeping for 30 days")
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case received, ok := <-signals:
			if !ok {
				return "", errors.New("complete signal channel closed")
			}
			if received.Err != nil {
				return "", received.Err
			}
			return "done", nil
		case <-time.After(30 * 24 * time.Hour):
		}
	}
}

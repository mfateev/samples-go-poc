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
		email := workflow.ExecuteActivity(ctx, SendEmail, "Sleeping for 30 days").ToChannel()
		timer := time.After(30 * 24 * time.Hour)
		// Finish both waits before starting the next iteration. Setting a
		// consumed channel to nil disables that arm of select.
		for email != nil || timer != nil {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case result := <-email:
				// SendEmail returns only error, so there is no value to extract.
				if result.Err != nil {
					return "", result.Err
				}
				email = nil
			case received, ok := <-signals:
				if !ok {
					return "", errors.New("complete signal channel closed")
				}
				if received.Err != nil {
					return "", received.Err
				}
				return "done", nil
			case <-timer:
				timer = nil
			}
		}
	}
}

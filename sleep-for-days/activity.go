package sleepfordays

import (
	"context"

	"go.temporal.io/sdk/activity"
)

// SendEmail is the host activity used by the upstream sleep-for-days sample.
func SendEmail(ctx context.Context, message string) error {
	activity.GetLogger(ctx).Info("Sending email", "message", message)
	return nil
}

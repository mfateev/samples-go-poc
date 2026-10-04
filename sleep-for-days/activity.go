package sleepfordays

import (
	"context"

	"go.temporal.io/sdk/activity"
)

// SendEmail is the host activity used by the upstream sleep-for-days sample.
func SendEmail(ctx context.Context, message []byte) ([]byte, error) {
	activity.GetLogger(ctx).Info("Sending email", "message", string(message))
	return nil, nil
}

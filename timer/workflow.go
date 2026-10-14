package timer

import (
	"context"
	"math/rand"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
	"go.temporal.io/sdk/activity"
)

// SampleTimerWorkflow workflow definition
//
//temporal:workflow
func SampleTimerWorkflow(ctx context.Context, processingTimeThreshold time.Duration) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Second})
	processing := workflow.ExecuteActivity(ctx, OrderProcessingActivity)
	timer := time.NewTimer(processingTimeThreshold)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-processing.ToChannel():
		return processing.Get(ctx, nil)
	case <-timer.C:
		if err := workflow.ExecuteActivity(ctx, SendEmailActivity).Get(ctx, nil); err != nil {
			return err
		}
	}
	return processing.Get(ctx, nil)
}

func OrderProcessingActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("OrderProcessingActivity processing started.")
	timeNeededToProcess := time.Second * time.Duration(rand.Intn(10))
	time.Sleep(timeNeededToProcess)
	logger.Info("OrderProcessingActivity done.", "duration", timeNeededToProcess)
	return nil
}

func SendEmailActivity(ctx context.Context) error {
	activity.GetLogger(ctx).Info("SendEmailActivity sending notification email as the process takes long time.")
	return nil
}

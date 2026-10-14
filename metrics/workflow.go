package metrics

import (
	"context"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
	"go.temporal.io/sdk/activity"
)

//temporal:workflow
func Workflow(ctx context.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Metrics workflow started.")

	scheduledTimeNanos := time.Now().UnixNano()
	_ = sleep(ctx, 500*time.Millisecond)
	err := workflow.ExecuteActivity(ctx, Activity, scheduledTimeNanos).Get(ctx, nil)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	logger.Info("Metrics workflow completed.")
	return nil
}

func Activity(ctx context.Context, scheduledTimeNanos int64) error {
	logger := activity.GetLogger(ctx)

	var err error
	metricsHandler := activity.GetMetricsHandler(ctx)
	metricsHandler = recordActivityStart(metricsHandler, "metrics.Activity", scheduledTimeNanos)
	startTime := time.Now()
	defer func() { recordActivityEnd(metricsHandler, startTime, err) }()

	time.Sleep(time.Second)
	logger.Info("Metrics reported.")
	return err
}

package opentelemetry

import (
	"context"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/activity"
)

//go:isolate
func Workflow(ctx context.Context, name string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorld workflow started", "name", name)

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})

	err := workflow.ExecuteActivity(ctx, Activity, name).Get(ctx, nil)

	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	logger.Info("HelloWorld workflow completed.")
	return nil
}

func Activity(ctx context.Context, name string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)

	// Get current span and add new attributes
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Bool("isTrue", true), attribute.String("stringAttr", "Ciao"))

	// Create a child span
	_, childSpan := span.TracerProvider().Tracer("temporal-sample").Start(ctx, "custom-span")
	time.Sleep(1 * time.Second)
	childSpan.End()

	time.Sleep(1 * time.Second)

	// Add an event to the current span
	span.AddEvent("Done Activity")

	return nil
}

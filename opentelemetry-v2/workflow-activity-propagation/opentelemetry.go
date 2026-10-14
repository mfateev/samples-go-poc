package workflowactivity

import (
	"context"
	"fmt"
	"time"

	temporalotel "github.com/mfateev/sdk-go-poc/contrib/opentelemetry-v2"
	"github.com/mfateev/sdk-go-poc/workflow"
	"go.opentelemetry.io/otel/trace"
)

const TaskQueueName = "workflow-activity-propagation"

// @@@SNIPSTART samples-go-opentelemetry-v2-application-spans
const instrumentationName = "github.com/temporalio/samples-go/opentelemetry-v2/workflow-activity-propagation"

//temporal:workflow
func Workflow(ctx context.Context, name string) (string, error) {
	tracer := temporalotel.NewSinkTracerProvider(1000).Tracer(instrumentationName)
	ctx, span := tracer.Start(ctx, "workflow-operation")
	defer span.End()

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})

	var result string
	if err := workflow.ExecuteActivity(ctx, Activity, name).Get(ctx, &result); err != nil {
		return "", err
	}

	return result, nil
}

func Activity(ctx context.Context, name string) (string, error) {
	_, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(instrumentationName).Start(ctx, "activity-operation")
	defer span.End()

	return fmt.Sprintf("Hello, %s!", name), nil
}

// @@@SNIPEND

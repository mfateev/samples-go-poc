package clientupdate

import (
	"fmt"

	"context"

	temporalotel "github.com/mfateev/sdk-go-poc/contrib/opentelemetry-v2"
	"github.com/mfateev/sdk-go-poc/workflow"
)

const (
	instrumentationName = "github.com/temporalio/samples-go/opentelemetry-v2/client-update-propagation"
	TaskQueueName       = "client-update-propagation"
	UpdateName          = "greet"
)

// @@@SNIPSTART samples-go-opentelemetry-v2-update-handler-span
//
//temporal:workflow
func Workflow(ctx context.Context) (string, error) {
	var result string
	updateCompleted := make(chan struct{})

	err := workflow.SetUpdateHandlerWithOptions(
		ctx,
		UpdateName,
		func(ctx context.Context, name string) (string, error) {
			_, span := temporalotel.NewSinkTracerProvider(1000).Tracer(instrumentationName).Start(ctx, "handle-update")
			defer span.End()

			result = fmt.Sprintf("Hello, %s!", name)
			close(updateCompleted)
			return result, nil
		},
		workflow.UpdateHandlerOptions{
			Validator: func(ctx context.Context, name string) error {
				_, span := temporalotel.NewSinkTracerProvider(1000).Tracer(instrumentationName).Start(ctx, "validate-update")
				defer span.End()

				if name == "" {
					return fmt.Errorf("name cannot be empty")
				}
				return nil
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("unable to register Update handler: %w", err)
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-updateCompleted:
	}
	if err := workflow.WaitHandlersFinished(ctx); err != nil {
		return "", err
	}
	return result, nil
}

// @@@SNIPEND

package handler

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/nexus-rpc/sdk-go/nexus"

	"errors"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"

	"github.com/mfateev/samples-go-poc/nexus/service"
)

// Use the NewWorkflowRunOperation constructor, which is the easiest way to expose a workflow as an operation.
// See alternatives at https://pkg.go.dev/go.temporal.io/sdk/temporalnexus.
var HelloOperation = temporalnexus.MustNewWorkflowRunOperationWithOptions(temporalnexus.WorkflowRunOperationOptions[service.HelloInput, service.HelloOutput]{Name: service.HelloOperationName, Handler: func(ctx context.Context, input service.HelloInput, options nexus.StartOperationOptions) (temporalnexus.WorkflowHandle[service.HelloOutput], error) {
	return temporalnexus.ExecuteUntypedWorkflow[service.HelloOutput](ctx, options, client.StartWorkflowOptions{ID: service.HelloWorkflowID(input)}, "HelloHandlerWorkflow", input)
}})

//temporal:workflow
func HelloHandlerWorkflow(ctx context.Context, input service.HelloInput) (service.HelloOutput, error) {
	// Sleep for a random duration to simulate some work
	duration := time.Duration(rand.IntN(5)) * time.Second
	if err := sleep(ctx, duration); err != nil {
		// Simulate some work after cancellation is requested
		sleepErr := err
		if errors.Is(err, context.Canceled) {
			ctx = context.WithoutCancel(ctx)
			duration := time.Duration(rand.IntN(5)) * time.Second
			if err := sleep(ctx, duration); err != nil {
				return service.HelloOutput{}, err
			}
		}
		return service.HelloOutput{}, sleepErr
	}

	switch input.Language {
	case service.EN:
		return service.HelloOutput{Message: "Hello " + input.Name + " 👋"}, nil
	case service.FR:
		return service.HelloOutput{Message: "Bonjour " + input.Name + " 👋"}, nil
	case service.DE:
		return service.HelloOutput{Message: "Hallo " + input.Name + " 👋"}, nil
	case service.ES:
		return service.HelloOutput{Message: "¡Hola! " + input.Name + " 👋"}, nil
	case service.TR:
		return service.HelloOutput{Message: "Merhaba " + input.Name + " 👋"}, nil
	}
	return service.HelloOutput{}, fmt.Errorf("unsupported language %q", input.Language)
}

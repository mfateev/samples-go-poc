// @@@SNIPSTART samples-go-nexus-handler
package handler

import (
	"context"
	"fmt"

	"github.com/nexus-rpc/sdk-go/nexus"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"

	"github.com/mfateev/samples-go-poc/nexus/service"
)

// NewSyncOperation is a meant for exposing simple RPC handlers.
var EchoOperation = nexus.NewSyncOperation(service.EchoOperationName, func(ctx context.Context, input service.EchoInput, options nexus.StartOperationOptions) (service.EchoOutput, error) {
	// Use temporalnexus.GetClient to get the client that the worker was initialized with to perform client calls
	// such as signaling, querying, and listing workflows. Implementations are free to make arbitrary calls to other
	// services or databases, or perform simple computations such as this one.
	return service.EchoOutput(input), nil
})

// Use the NewWorkflowRunOperation constructor, which is the easiest way to expose a workflow as an operation.
// See alternatives at https://pkg.go.dev/go.temporal.io/sdk/temporalnexus.
var HelloOperation = temporalnexus.MustNewWorkflowRunOperationWithOptions(temporalnexus.WorkflowRunOperationOptions[service.HelloInput, service.HelloOutput]{Name: service.HelloOperationName, Handler: func(ctx context.Context, input service.HelloInput, options nexus.StartOperationOptions) (temporalnexus.WorkflowHandle[service.HelloOutput], error) {
	return temporalnexus.ExecuteUntypedWorkflow[service.HelloOutput](ctx, options, client.StartWorkflowOptions{ID: service.HelloWorkflowID(input)}, "HelloHandlerWorkflow", input)
}})

//go:isolate
func HelloHandlerWorkflow(_ context.Context, input service.HelloInput) (service.HelloOutput, error) {
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

// @@@SNIPEND

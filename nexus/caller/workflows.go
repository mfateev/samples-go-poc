// @@@SNIPSTART samples-go-nexus-caller-workflow
package caller

import (
	"context"

	"github.com/mfateev/samples-go-poc/nexus/service"
	"github.com/mfateev/sdk-go-poc/workflow"
)

const (
	TaskQueue    = "my-caller-workflow-task-queue"
	endpointName = "my-nexus-endpoint-name"
)

//go:isolate
func EchoCallerWorkflow(ctx context.Context, message string) (string, error) {
	c := workflow.NewNexusClient(endpointName, service.HelloServiceName)

	fut := c.ExecuteOperation(ctx, service.EchoOperationName, service.EchoInput{Message: message}, workflow.NexusOperationOptions{})

	var res service.EchoOutput
	if err := fut.Get(ctx, &res); err != nil {
		return "", err
	}

	return res.Message, nil
}

//go:isolate
func HelloCallerWorkflow(ctx context.Context, name string, language service.Language) (string, error) {
	c := workflow.NewNexusClient(endpointName, service.HelloServiceName)

	fut := c.ExecuteOperation(ctx, service.HelloOperationName, service.HelloInput{Name: name, Language: language}, workflow.NexusOperationOptions{})
	var res service.HelloOutput

	// Optionally wait for the operation to be started. NexusOperationExecution will contain the operation token in
	// case this operation is asynchronous, which is a handle that can be used to perform additional actions like
	// cancelling an operation.
	var exec workflow.NexusOperationExecution
	if err := fut.GetNexusOperationExecution().Get(ctx, &exec); err != nil {
		return "", err
	}
	if err := fut.Get(ctx, &res); err != nil {
		return "", err
	}

	return res.Message, nil
}

// @@@SNIPEND

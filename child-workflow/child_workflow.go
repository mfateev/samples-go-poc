package child_workflow

import (
	"context"

	"github.com/mfateev/sdk-go-poc/workflow"
)

// @@@SNIPSTART samples-go-child-workflow-example-child-workflow-definition
// SampleChildWorkflow is a Workflow Definition
//
//go:isolate
func SampleChildWorkflow(ctx context.Context, name string) (string, error) {
	logger := workflow.GetLogger(ctx)
	greeting := "Hello " + name + "!"
	logger.Info("Child workflow execution: " + greeting)
	return greeting, nil
}

// @@@SNIPEND

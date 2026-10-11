package datadog

import (
	"context"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
	"go.temporal.io/sdk/activity"
)

//go:isolate
func Workflow(ctx context.Context, name string) error {
	workflow.GetLogger(ctx).Info("Executing Workflow.", "name", name)
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID: "DATADOG-CHILD-WORKFLOW-ID",
	}
	ctx = workflow.WithChildWorkflowOptions(ctx, cwo)
	return workflow.ExecuteChildWorkflow(ctx, ChildWorkflow, name).Get(ctx, nil)
}

//go:isolate
func ChildWorkflow(ctx context.Context, name string) error {
	workflow.GetLogger(ctx).Info("Executing ChildWorkflow.", "name", name)
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	return workflow.ExecuteActivity(ctx, Activity, name).Get(ctx, nil)
}

func Activity(ctx context.Context, name string) error {
	activity.GetLogger(ctx).Info("Executing Activity.", "name", name)
	time.Sleep(time.Second)
	return nil
}

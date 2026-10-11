package cancellation

import (
	"context"
	"errors"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
)

// @@@SNIPSTART samples-go-cancellation-workflow-definition
// YourWorkflow is a Workflow Definition that shows how it can be canceled.
//
//go:isolate
func YourWorkflow(ctx context.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		HeartbeatTimeout:    5 * time.Second,
		WaitForCancellation: true,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("cancel workflow started")
	var a *Activities // Used to call Activities by function pointer
	defer func() {

		if !errors.Is(ctx.Err(), context.Canceled) {
			return
		}

		// When the Workflow is canceled, it has to get a new disconnected context to execute any Activities
		// Specify activity options for cleanup activity
		aoCleanup := workflow.ActivityOptions{
			StartToCloseTimeout: 2 * time.Second,
		}
		newCtx := context.WithoutCancel(ctx)
		newCtx = workflow.WithActivityOptions(newCtx, aoCleanup)
		err := workflow.ExecuteActivity(newCtx, a.CleanupActivity).Get(newCtx, nil)
		if err != nil {
			logger.Error("CleanupActivity failed", "Error", err)
		}

	}()

	var result string
	err := workflow.ExecuteActivity(ctx, a.ActivityToBeCanceled).Get(ctx, &result)
	if err != nil {
		return err
	}

	logger.Info("Workflow Execution complete.")
	return nil
}

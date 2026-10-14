package largeeventhistory

import (
	"errors"
	"fmt"
	"time"

	"context"

	"github.com/mfateev/sdk-go-poc/workflow"
)

// LargeEventHistoryWorkflow workflow definition
//
//temporal:workflow
func LargeEventHistoryWorkflow(ctx context.Context, LengthOfHistory int, WillFailOrNot bool) (err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	countActivities := LengthOfHistory / 6
	for range countActivities {
		if err := workflow.ExecuteActivity(ctx, Activity).Get(ctx, nil); err != nil {
			return fmt.Errorf("activity failed: %w", err)
		}
	}

	if WillFailOrNot {
		return errors.New("intentional workflow failure due to WillFailOrNot parameter")
	}
	return nil
}

func Activity(ctx context.Context) error {
	return nil
}

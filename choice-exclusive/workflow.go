// The upstream exclusive-choice workflow, adapted to run in an isolate.
package choiceexclusive

import (
	"context"
	"fmt"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
)

const (
	OrderChoiceApple  = "apple"
	OrderChoiceBanana = "banana"
	OrderChoiceCherry = "cherry"
	OrderChoiceOrange = "orange"
)

//go:isolate
func ExclusiveChoice(ctx context.Context) error {
	// A nil receiver is only a method identifier. Temporal invokes the
	// registered host OrderActivities instance, with its configured choices.
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Second})
	var orders *OrderActivities
	var choice string
	err := workflow.ExecuteActivity(ctx, orders.GetOrder).Get(ctx, &choice)
	if err != nil {
		return err
	}
	var selected func(context.Context, string) error
	switch choice {
	case OrderChoiceApple:
		selected = orders.OrderApple
	case OrderChoiceBanana:
		selected = orders.OrderBanana
	case OrderChoiceCherry:
		selected = orders.OrderCherry
	case OrderChoiceOrange:
		selected = orders.OrderOrange
	default:
		return fmt.Errorf("unknown order choice: %v", choice)
	}
	return workflow.ExecuteActivity(ctx, selected, choice).Get(ctx, nil)
}

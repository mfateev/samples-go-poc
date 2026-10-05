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
	var orders *OrderActivities
	choice, err := workflow.ExecuteActivityNoInput(ctx, orders.GetOrder, 10*time.Second)
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
	return workflow.ExecuteActivityError(ctx, selected, 10*time.Second, choice)
}

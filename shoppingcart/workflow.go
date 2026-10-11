package shoppingcart

import (
	"context"
	"fmt"

	"github.com/mfateev/sdk-go-poc/workflow"
)

var (
	UpdateName    = "shopping-cart"
	TaskQueueName = "shopping-cart-tq"
)

type CartState struct {
	Items map[string]int // itemID -> quantity
}

//go:isolate
func CartWorkflow(ctx context.Context, cart *CartState) error {
	if cart == nil {
		cart = &CartState{make(map[string]int)}
	}
	if cart.Items == nil {
		cart.Items = make(map[string]int)
	}
	logger := workflow.GetLogger(ctx)
	changed := make(chan struct{}, 1)

	if err := workflow.SetUpdateHandlerWithOptions(ctx, UpdateName, func(ctx context.Context, actionType string, itemID string) (*CartState, error) {
		defer func() {
			select {
			case changed <- struct{}{}:
			default:
			}
		}()
		logger.Info("Received update,", actionType, itemID)
		switch actionType {
		case "add":
			cart.Items[itemID] += 1
		case "remove":
			cart.Items[itemID] -= 1
			if cart.Items[itemID] <= 0 {
				delete(cart.Items, itemID)
			}
		case "list":
		default:
			logger.Error("Unsupported action type.")
		}

		return cart, nil
	}, workflow.UpdateHandlerOptions{
		Validator: func(ctx context.Context, actionType string, itemID string) error {
			switch actionType {
			case "add", "remove":
				if itemID == "" {
					return fmt.Errorf("itemID must be specified for add or remove actionType")
				}
			case "list":
				if itemID != "" {
					logger.Warn("ItemID not needed for \"list\" actionType.")
				}
			default:
				return fmt.Errorf("unsupported action type: %s", actionType)
			}
			return nil
		},
	}); err != nil {
		return err
	}

	checkout := workflow.GetSignalChannel[any](ctx, "checkout")
	for {
		if workflow.GetInfo(ctx).GetContinueAsNewSuggested() {
			if err := workflow.WaitHandlersFinished(ctx); err != nil {
				return err
			}
			return workflow.NewContinueAsNewError(ctx, CartWorkflow, cart)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-changed:
		case <-checkout:
			return workflow.WaitHandlersFinished(ctx)
		}
	}
}

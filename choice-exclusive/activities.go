package choiceexclusive

import (
	"context"
	"fmt"
	"math/rand"
)

// OrderActivities configures the choices offered by the host activities.
type OrderActivities struct {
	OrderChoices []string
}

func (a *OrderActivities) GetOrder(_ context.Context) (string, error) {
	order := a.OrderChoices[rand.Intn(len(a.OrderChoices))]
	fmt.Printf("Order is for %s\n", order)
	return order, nil
}

func (a *OrderActivities) OrderApple(_ context.Context, choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func (a *OrderActivities) OrderBanana(_ context.Context, choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func (a *OrderActivities) OrderCherry(_ context.Context, choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func (a *OrderActivities) OrderOrange(_ context.Context, choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

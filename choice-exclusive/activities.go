package choiceexclusive

import (
	"fmt"
	"math/rand"
)

// OrderActivities configures the choices offered by the host activities.
type OrderActivities struct {
	OrderChoices []string
}

func (a *OrderActivities) GetOrder() (string, error) {
	order := a.OrderChoices[rand.Intn(len(a.OrderChoices))]
	fmt.Printf("Order is for %s\n", order)
	return order, nil
}

func (a *OrderActivities) OrderApple(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func (a *OrderActivities) OrderBanana(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func (a *OrderActivities) OrderCherry(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func (a *OrderActivities) OrderOrange(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

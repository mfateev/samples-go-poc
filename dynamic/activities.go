package dynamic

import (
	"context"
	"fmt"
)

type Activities struct{}

// Get Name Activity.
func (a *Activities) GetName(ctx context.Context) (string, error) {
	return "Temporal", nil
}

// Get Greeting Activity.
func (a *Activities) GetGreeting(ctx context.Context) (string, error) {
	return "Hello", nil
}

// Say Greeting Activity.
func (a *Activities) SayGreeting(ctx context.Context, greeting string, name string) (string, error) {
	result := fmt.Sprintf("Greeting: %s %s!\n", greeting, name)
	return result, nil
}

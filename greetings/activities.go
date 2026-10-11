package greetings

import (
	"context"
	"fmt"
	// @@@SNIPSTART samples-go-dependency-sharing-activities
)

type Activities struct {
	Name     string
	Greeting string
}

// GetGreeting Activity.
func (a *Activities) GetGreeting(ctx context.Context) (string, error) {
	return a.Greeting, nil
}

// @@@SNIPEND

// GetName Activity.
func (a *Activities) GetName(ctx context.Context) (string, error) {
	return a.Name, nil
}

// SayGreeting Activity.
func (a *Activities) SayGreeting(ctx context.Context, greeting string, name string) (string, error) {
	result := fmt.Sprintf("Greeting: %s %s!\n", greeting, name)
	return result, nil
}

package greetingslocal

import (
	"context"
	"fmt"
)

type Activities struct {
	Name     string
	Greeting string
}

// GetName Activity.
func (a *Activities) GetName(ctx context.Context) (string, error) {
	return a.Name, nil
}

// GetGreeting Activity.
func (a *Activities) GetGreeting(ctx context.Context) (string, error) {
	return a.Greeting, nil
}

// SayGreeting Activity.
func (a *Activities) SayGreeting(ctx context.Context, greeting string, name string) (string, error) {
	result := fmt.Sprintf("Greeting: %s %s!\n", greeting, name)
	return result, nil
}

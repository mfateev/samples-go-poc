package goroutines

import "context"

// Greet runs in the host worker, outside the workflow isolate.
func Greet(_ context.Context, name string) (string, error) {
	return "Hello " + name + "!", nil
}

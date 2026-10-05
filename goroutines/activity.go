package goroutines

import "context"

// Greet runs in the host worker, outside the workflow isolate.
func Greet(_ context.Context, name []byte) ([]byte, error) {
	return []byte("Hello " + string(name) + "!"), nil
}

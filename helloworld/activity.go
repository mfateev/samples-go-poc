package helloworld

import "context"

// Activity returns the same greeting as the upstream helloworld sample.
func Activity(_ context.Context, name string) (string, error) {
	return "Hello " + name + "!", nil
}

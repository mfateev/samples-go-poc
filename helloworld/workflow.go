// The upstream helloworld workflow, adapted to run as a named isolate function.
package helloworld

import (
	"context"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
)

//go:isolate
func HelloWorld(ctx context.Context, name string) (string, error) {
	greeting, err := workflow.ExecuteActivity(ctx, Activity, 10*time.Second, name)
	return greeting, err
}

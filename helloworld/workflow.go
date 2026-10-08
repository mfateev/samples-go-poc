// The upstream helloworld workflow, adapted to run as a named isolate function.
package helloworld

import (
	"context"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
)

//go:isolate
func HelloWorld(ctx context.Context, name string) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Second})
	var greeting string
	err := workflow.ExecuteActivity(ctx, Activity, name).Get(ctx, &greeting)
	return greeting, err
}

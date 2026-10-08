// This example uses ordinary Go goroutines and synchronization in a workflow.
package goroutines

import (
	"context"
	"sync"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
)

type greetingResult struct {
	index    int
	greeting string
	err      error
}

// GreetAll starts one activity per name concurrently. Results retain input
// order even when Temporal delivers activity completions in a different order.
//
//go:isolate
func GreetAll(ctx context.Context, names []string) ([]string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Second})
	results := make(chan greetingResult)
	var pending sync.WaitGroup
	for index, name := range names {
		pending.Add(1)
		go func() {
			defer pending.Done()
			outcome := <-workflow.ExecuteActivity(ctx, Greet, name).ToChannel()
			var greeting string
			err := outcome.Err
			if err == nil {
				err = outcome.Value.Get(&greeting)
			}
			results <- greetingResult{index: index, greeting: greeting, err: err}
		}()
	}
	go func() {
		pending.Wait()
		close(results)
	}()
	greetings := make([]string, len(names))
	for result := range results {
		if result.err != nil {
			return nil, result.err
		}
		greetings[result.index] = result.greeting
	}
	return greetings, nil
}

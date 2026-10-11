package goroutine

import (
	"fmt"
	"math/rand"
	"time"

	"context"

	"github.com/mfateev/sdk-go-poc/workflow"
)

/**
* This sample runs activity sequences concurrently with native goroutines.
* Each goroutine owns its result; a channel collects results in input order.
 */

// SampleGoroutineWorkflow workflow definition
//
//go:isolate
func SampleGoroutineWorkflow(ctx context.Context, parallelism int) ([]string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Second})
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	type result struct {
		index int
		value string
		err   error
	}
	completed := make(chan result, parallelism)
	for i := range parallelism {
		go func() {
			var first, second string
			err := workflow.ExecuteActivity(ctx, Step1, fmt.Sprint(i)).Get(ctx, &first)
			if err == nil {
				err = workflow.ExecuteActivity(ctx, Step2, first).Get(ctx, &second)
			}
			completed <- result{i, second, err}
		}()
	}
	results := make([]string, parallelism)
	for range parallelism {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case r := <-completed:
			if r.err != nil {
				return nil, r.err
			}
			results[r.index] = r.value
		}
	}
	return results, nil
}

func Step1(ctx context.Context, input string) (output string, err error) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	return input + ", Step1", nil
}

func Step2(ctx context.Context, input string) (output string, err error) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	return input + ", Step2", nil
}

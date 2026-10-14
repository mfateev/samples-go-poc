package splitmerge_selector

import (
	"context"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
	"go.temporal.io/sdk/activity"
)

/**
 * This sample workflow demonstrates how to execute multiple activities in parallel and merge their results using futures.
 * A native channel collects futures as they complete. It allows processing them as soon as they become ready. See `split-merge-future` sample
 * to see how to process them without Selector in the order of activity invocation instead.
 */

// ChunkResult contains the activity result for this sample
type ChunkResult struct {
	NumberOfItemsInChunk int
	SumInChunk           int
}

// SampleSplitMergeSelectorWorkflow workflow definition
//
//temporal:workflow
func SampleSplitMergeSelectorWorkflow(ctx context.Context, workerCount int) (ChunkResult, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Second})
	type completion struct {
		result ChunkResult
		err    error
	}
	completed := make(chan completion, workerCount)
	for i := range workerCount {
		f := workflow.ExecuteActivity(ctx, ChunkProcessingActivity, i+1)
		go func() { var r ChunkResult; err := f.Get(ctx, &r); completed <- completion{r, err} }()
	}
	var total ChunkResult
	for range workerCount {
		select {
		case <-ctx.Done():
			return ChunkResult{}, ctx.Err()
		case c := <-completed:
			if c.err != nil {
				return ChunkResult{}, c.err
			}
			total.NumberOfItemsInChunk += c.result.NumberOfItemsInChunk
			total.SumInChunk += c.result.SumInChunk
		}
	}
	return total, nil
}

func ChunkProcessingActivity(ctx context.Context, chunkID int) (result ChunkResult, err error) {

	// some fake processing logic here
	numberOfItemsInChunk := chunkID
	sumInChunk := chunkID * chunkID

	activity.GetLogger(ctx).Info("Chunk processed", "chunkID", chunkID)
	return ChunkResult{numberOfItemsInChunk, sumInChunk}, nil
}

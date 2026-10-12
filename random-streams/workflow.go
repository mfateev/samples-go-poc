package random_streams

import (
	"context"
	"errors"

	"github.com/mfateev/sdk-go-poc/workflow"
)

type Input struct {
	Count      int
	AuditDraws int
}

type Result struct {
	Choices      []int
	JitterMillis []int
}

// Plan uses stable names to separate choices from jitter and optional audit draws.
//
//go:isolate
func Plan(ctx context.Context, input Input) (Result, error) {
	if input.Count < 1 || input.Count > 100 || input.AuditDraws < 0 || input.AuditDraws > 1000 {
		return Result{}, errors.New("count must be 1..100 and audit draws must be 0..1000")
	}
	choices := workflow.GetRandomStream(ctx, "choices")
	jitter := workflow.GetRandomStream(ctx, "jitter")
	audit := workflow.GetRandomStream(ctx, "audit")
	for range input.AuditDraws {
		_ = audit.Uint64()
	}
	result := Result{Choices: make([]int, input.Count), JitterMillis: make([]int, input.Count)}
	for i := range result.Choices {
		result.Choices[i] = choices.IntN(6) + 1
		result.JitterMillis[i] = jitter.IntN(1000)
	}
	if err := workflow.SetQueryHandler(ctx, "next", func() (int, error) {
		// Queries use a scratch copy, including through a captured generator.
		return choices.IntN(6) + 1, nil
	}); err != nil {
		return Result{}, err
	}
	return result, nil
}

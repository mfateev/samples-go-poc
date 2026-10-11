package update

import (
	"fmt"

	"context"

	"github.com/mfateev/sdk-go-poc/workflow"
)

const (
	FetchAndAdd = "fetch_and_add"
	Done        = "done"
)

//go:isolate
func Counter(ctx context.Context) (int, error) {
	log := workflow.GetLogger(ctx)
	counter := 0

	if err := workflow.SetUpdateHandlerWithOptions(
		ctx,
		FetchAndAdd,
		func(ctx context.Context, i int) (int, error) {
			tmp := counter
			counter += i
			log.Info("counter updated", "addend", i, "new-value", counter)
			return tmp, nil
		},
		workflow.UpdateHandlerOptions{Validator: nonNegative},
	); err != nil {
		return 0, err
	}

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-workflow.GetSignalChannel[any](ctx, Done):
	}
	if err := workflow.WaitHandlersFinished(ctx); err != nil {
		return 0, err
	}
	return counter, ctx.Err()
}

func nonNegative(ctx context.Context, i int) error {
	log := workflow.GetLogger(ctx)
	if i < 0 {
		log.Debug("Rejecting negative update", "addend", i)
		return fmt.Errorf("addend must be non-negative (%v)", i)
	}
	log.Debug("Accepting update", "addend", i)
	return nil
}

package await_signals

import (
	"context"
	"fmt"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
	"go.temporal.io/sdk/temporal"
)

var SignalToSignalTimeout = 30 * time.Second
var FromFirstSignalTimeout = 60 * time.Second

// AwaitSignalsWorkflow processes signals in order, retaining early arrivals.
//
//go:isolate
func AwaitSignalsWorkflow(ctx context.Context) error {
	channels := []<-chan any{workflow.GetSignalChannel[any](ctx, "Signal1"), workflow.GetSignalChannel[any](ctx, "Signal2"), workflow.GetSignalChannel[any](ctx, "Signal3")}
	received := [3]bool{}
	var first time.Time
	next := 0
	var deadline <-chan time.Time
	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()
	for next < 3 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return temporal.NewApplicationError(fmt.Sprintf("Timed out waiting for signal%d", next+1), "timeout")
		case <-channels[0]:
			received[0] = true
			channels[0] = nil
		case <-channels[1]:
			received[1] = true
			channels[1] = nil
		case <-channels[2]:
			received[2] = true
			channels[2] = nil
		}
		if first.IsZero() {
			first = time.Now()
		}
		for next < 3 && received[next] {
			workflow.GetLogger(ctx).Info(fmt.Sprintf("Signal%d Processed", next+1))
			next++
			if next == 3 {
				return nil
			}
			remaining := FromFirstSignalTimeout - time.Since(first)
			if remaining <= 0 {
				return temporal.NewApplicationError("FromFirstSignalTimeout", "timeout")
			}
			duration := min(SignalToSignalTimeout, remaining)
			if timer == nil {
				timer = time.NewTimer(duration)
			} else {
				timer.Reset(duration)
			}
			deadline = timer.C
		}
	}
	return nil
}

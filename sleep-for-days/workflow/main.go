// The upstream sleep-for-days workflow, adapted to use native Go channels,
// select, and channel-backed futures inside an isolate.
package main

import (
	"errors"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
)

const defaultInterval = 30 * 24 * time.Hour

type future struct{ ready <-chan error }

// The SDK's blocking operations become futures when an owned goroutine calls
// them. The buffered result channel lets the goroutine finish after the host
// replies, even if another select case won first.
func startFuture(call func() error) future {
	ready := make(chan error, 1)
	go func() { ready <- call() }()
	return future{ready: ready}
}

type signalResult struct {
	signal workflow.Signal
	err    error
}

func receiveSignals() <-chan signalResult {
	results := make(chan signalResult, 1)
	go func() {
		for {
			signal, err := workflow.NextSignal()
			results <- signalResult{signal: signal, err: err}
			if err != nil {
				return
			}
		}
	}()
	return results
}

func main() {
	result, err := run()
	_ = workflow.Complete(result, err)
}

func run() ([]byte, error) {
	input, err := workflow.Input()
	if err != nil {
		return nil, err
	}
	interval := defaultInterval
	if len(input) != 0 {
		interval, err = time.ParseDuration(string(input))
		if err != nil {
			return nil, err
		}
		if interval <= 0 {
			return nil, errors.New("sleep interval must be positive")
		}
	}

	signals := receiveSignals()
	for {
		activity := startFuture(func() error {
			_, err := workflow.ExecuteActivity("SendEmail", []byte("Sleeping for "+interval.String()), 10*time.Second)
			return err
		})
		timer := startFuture(func() error { return workflow.Sleep(interval) })
		for timer.ready != nil {
			select {
			case received := <-signals:
				if received.err != nil {
					return nil, received.err
				}
				if received.signal.Name == "complete" {
					return []byte("done"), nil
				}
			case err := <-activity.ready:
				if err != nil {
					return nil, err
				}
				activity.ready = nil
			case err := <-timer.ready:
				if err != nil {
					return nil, err
				}
				timer.ready = nil
			}
		}
	}
}

// The upstream sleep-for-days workflow, adapted to use native Go channels and
// select inside an isolate.
package main

import (
	"errors"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
)

const defaultInterval = 30 * 24 * time.Hour

func init() {
	workflow.RegisterTyped("SleepForDays", SleepForDays)
}

func main() {
	_ = workflow.Run()
}

func SleepForDays(input string) (string, error) {
	interval := defaultInterval
	if input != "" {
		var err error
		interval, err = time.ParseDuration(input)
		if err != nil {
			return "", err
		}
		if interval <= 0 {
			return "", errors.New("sleep interval must be positive")
		}
	}

	signals := workflow.GetSignalChannel("complete")
	for {
		activity := workflow.ExecuteActivityAsync("SendEmail", []byte("Sleeping for "+interval.String()), 10*time.Second)
		timer := time.After(interval)
		for timer != nil {
			select {
			case received, ok := <-signals:
				if !ok {
					return "", errors.New("complete signal channel closed")
				}
				if received.Err != nil {
					return "", received.Err
				}
				return "done", nil
			case completed, ok := <-activity:
				if !ok {
					return "", errors.New("activity result channel closed")
				}
				if completed.Err != nil {
					return "", completed.Err
				}
				activity = nil
			case <-timer:
				timer = nil
			}
		}
	}
}

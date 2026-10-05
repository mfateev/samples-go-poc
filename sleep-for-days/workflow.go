// The upstream sleep-for-days workflow, adapted to use native Go channels and
// select inside an isolate.
package sleepfordays

import (
	"errors"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
)

//go:isolate
func SleepForDays() (string, error) {
	signals := workflow.GetSignalChannel("complete")
	for {
		// The upstream sample schedules the email without awaiting its future.
		// Completion or failure of the email does not control this workflow.
		_ = workflow.ExecuteActivityAsyncByName[struct{}]("SendEmail", 10*time.Second, "Sleeping for 30 days")
		select {
		case received, ok := <-signals:
			if !ok {
				return "", errors.New("complete signal channel closed")
			}
			if received.Err != nil {
				return "", received.Err
			}
			return "done", nil
		case <-time.After(30 * 24 * time.Hour):
		}
	}
}

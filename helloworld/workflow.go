// The upstream helloworld workflow, adapted to run as a named isolate function.
package helloworld

import (
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
)

//go:isolate
func HelloWorld(name string) (string, error) {
	greeting, err := workflow.ExecuteActivityWithContext(Activity, 10*time.Second, name)
	return greeting, err
}

// Replay a completed sample from a CLI-exported Temporal history.
package main

import (
	"fmt"
	"os"

	choice "github.com/mfateev/samples-go-poc/choice-exclusive"
	hello "github.com/mfateev/samples-go-poc/helloworld"
	sleep "github.com/mfateev/samples-go-poc/sleep-for-days"
	"github.com/mfateev/sdk-go-poc/worker"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: replay <workflow-type> <history.json>")
		os.Exit(2)
	}
	functions := map[string]any{
		"HelloWorld":      hello.HelloWorld,
		"ExclusiveChoice": choice.ExclusiveChoice,
		"SleepForDays":    sleep.SleepForDays,
	}
	fn, ok := functions[os.Args[1]]
	if !ok {
		fmt.Fprintln(os.Stderr, "unknown workflow type:", os.Args[1])
		os.Exit(2)
	}
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflow(fn)
	if err := replayer.ReplayWorkflowHistoryFromJSONFile(nil, os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("replay passed:", os.Args[1])
}

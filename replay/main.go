// Replay a completed sample from a CLI-exported Temporal history.
package main

import (
	"fmt"
	"os"
	"slices"

	choice "github.com/mfateev/samples-go-poc/choice-exclusive"
	"github.com/mfateev/samples-go-poc/goroutines"
	hello "github.com/mfateev/samples-go-poc/helloworld"
	sleep "github.com/mfateev/samples-go-poc/sleep-for-days"
	"github.com/mfateev/sdk-go-poc/worker"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: replay <workflow-type> <history.json>")
		os.Exit(2)
	}
	functions := map[string]any{
		"HelloWorld":      hello.HelloWorld,
		"GreetAll":        goroutines.GreetAll,
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
	if os.Args[1] == "GreetAll" {
		if err := compareGreetings(replayer, os.Args[2]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	fmt.Println("replay passed:", os.Args[1])
}

func compareGreetings(replayer worker.WorkflowReplayer, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	history, err := client.HistoryFromJSON(f, client.HistoryJSONOptions{})
	if err != nil {
		return err
	}
	if len(history.Events) == 0 {
		return fmt.Errorf("empty history")
	}
	completed := history.Events[len(history.Events)-1].GetWorkflowExecutionCompletedEventAttributes()
	if completed == nil {
		return fmt.Errorf("history has no completed workflow result")
	}
	var expected []string
	if err := converter.GetDefaultDataConverter().FromPayloads(completed.Result, &expected); err != nil {
		return err
	}
	getter, ok := replayer.(interface{ GetWorkflowResult(string, any) error })
	if !ok {
		return fmt.Errorf("replayer does not expose workflow results")
	}
	var actual []string
	if err := getter.GetWorkflowResult("", &actual); err != nil {
		return err
	}
	if !slices.Equal(actual, expected) {
		return fmt.Errorf("greetings changed on replay: %v, history has %v", actual, expected)
	}
	return nil
}

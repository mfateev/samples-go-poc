// The upstream helloworld workflow, adapted to run as a named isolate function.
package main

import (
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
)

func init() {
	workflow.RegisterTyped("HelloWorld", HelloWorld)
}

func main() {
	_ = workflow.Run()
}

func HelloWorld(name string) (string, error) {
	greeting, err := workflow.ExecuteActivity("HelloWorldActivity", []byte(name), 10*time.Second)
	return string(greeting), err
}

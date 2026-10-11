package main

import (
	"log"

	largeeventhistory "github.com/mfateev/samples-go-poc/temporal-fixtures/large-event-history"
	isolateworker "github.com/mfateev/sdk-go-poc/worker"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := isolateworker.New(c, "largeeventhistory", worker.Options{})

	w.RegisterWorkflow(largeeventhistory.LargeEventHistoryWorkflow)
	w.RegisterActivity(largeeventhistory.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

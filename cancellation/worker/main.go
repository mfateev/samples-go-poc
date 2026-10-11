package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/mfateev/samples-go-poc/cancellation"
	isolateworker "github.com/mfateev/sdk-go-poc/worker"
)

// @@@SNIPSTART samples-go-cancellation-worker-starter
func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := isolateworker.New(c, "cancel-activity", worker.Options{})

	w.RegisterWorkflow(cancellation.YourWorkflow)
	w.RegisterActivity(&cancellation.Activities{})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

// @@@SNIPEND

package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	cw "github.com/mfateev/samples-go-poc/child-workflow-continue-as-new"
	isolateworker "github.com/mfateev/sdk-go-poc/worker"
)

// @@@SNIPSTART samples-go-cw-cas-worker-starter
func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := isolateworker.New(c, "child-workflow-continue-as-new", worker.Options{})

	w.RegisterWorkflow(cw.SampleParentWorkflow)
	w.RegisterWorkflow(cw.SampleChildWorkflow)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

// @@@SNIPEND

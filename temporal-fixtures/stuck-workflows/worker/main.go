package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	stuckworkflows "github.com/mfateev/samples-go-poc/temporal-fixtures/stuck-workflows"
	isolateworker "github.com/mfateev/sdk-go-poc/worker"
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

	w := isolateworker.New(c, "stuck-workflows", worker.Options{})

	w.RegisterWorkflow(stuckworkflows.StuckWorkflow)
	w.RegisterActivity(stuckworkflows.StuckWorkflowActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

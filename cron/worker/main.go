package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/mfateev/samples-go-poc/cron"
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

	w := isolateworker.New(c, "cron", worker.Options{})

	w.RegisterWorkflow(cron.SampleCronWorkflow)
	w.RegisterActivity(cron.DoSomething)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

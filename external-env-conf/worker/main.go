package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	externalenvconf "github.com/mfateev/samples-go-poc/external-env-conf"
	isolateworker "github.com/mfateev/sdk-go-poc/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(externalenvconf.LoadProfile())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := isolateworker.New(c, "hello-world", worker.Options{})

	w.RegisterWorkflow(externalenvconf.Workflow)
	w.RegisterActivity(externalenvconf.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

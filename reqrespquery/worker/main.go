package main

import (
	"log"

	"github.com/mfateev/samples-go-poc/reqrespquery"
	isolateworker "github.com/mfateev/sdk-go-poc/worker"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := isolateworker.New(c, "reqrespquery", worker.Options{})

	w.RegisterWorkflow(reqrespquery.UppercaseWorkflow)
	w.RegisterActivity(reqrespquery.UppercaseActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

package main

import (
	"log"

	grpcproxy "github.com/mfateev/samples-go-poc/grpc-proxy"
	isolateworker "github.com/mfateev/sdk-go-poc/worker"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: "localhost:8081",
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := isolateworker.New(c, "grpcproxy", worker.Options{})

	w.RegisterWorkflow(grpcproxy.Workflow)
	w.RegisterActivity(grpcproxy.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

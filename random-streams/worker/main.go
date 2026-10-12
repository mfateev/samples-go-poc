package main

import (
	"log"
	"os"

	example "github.com/mfateev/samples-go-poc/random-streams"
	"github.com/mfateev/sdk-go-poc/worker"
	"go.temporal.io/sdk/client"
)

func main() {
	address := os.Getenv("TEMPORAL_ADDRESS")
	if address == "" {
		address = client.DefaultHostPort
	}
	c, err := client.Dial(client.Options{HostPort: address})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	w := worker.New(c, "random-streams-poc", worker.Options{})
	w.RegisterWorkflow(example.Plan)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"log"
	"os"

	"github.com/mfateev/samples-go-poc/goroutines"
	"github.com/mfateev/sdk-go-poc/worker"
	"go.temporal.io/sdk/activity"
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
	w := worker.New(c, "goroutines-poc", worker.Options{})
	w.RegisterWorkflow(goroutines.GreetAll)
	w.RegisterActivityWithOptions(goroutines.Greet, activity.RegisterOptions{Name: "Greet"})
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatal(err)
	}
}

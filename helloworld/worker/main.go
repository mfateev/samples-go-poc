package main

import (
	"log"
	"os"

	"github.com/mfateev/samples-go-poc/helloworld"
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
	w := worker.New(c, "hello-world-poc", worker.Options{})
	w.RegisterWorkflow(helloworld.HelloWorld)
	w.RegisterActivityWithOptions(helloworld.Activity, activity.RegisterOptions{Name: "HelloWorldActivity"})
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatal(err)
	}
}

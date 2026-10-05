package main

import (
	"log"
	"os"

	sleepfordays "github.com/mfateev/samples-go-poc/sleep-for-days"
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
	w := worker.New(c, "sleep-for-days-poc", worker.Options{})
	w.RegisterWorkflow(sleepfordays.SleepForDays)
	w.RegisterActivityWithOptions(sleepfordays.SendEmail, activity.RegisterOptions{Name: "SendEmail"})
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatal(err)
	}
}

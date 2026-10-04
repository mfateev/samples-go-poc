package main

import (
	"isolate"
	"log"
	"os"

	sleepfordays "github.com/mfateev/samples-go-poc/sleep-for-days"
	"github.com/mfateev/sdk-go-poc/temporalbridge"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	program, ok := isolate.LookupProgram("sleep-for-days-poc")
	if !ok {
		log.Fatal("missing sleep-for-days-poc; build with -isolate-dir=./sleep-for-days/workflow")
	}
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
	temporalbridge.Register(w, "SleepForDays", program)
	w.RegisterActivityWithOptions(sleepfordays.SendEmail, activity.RegisterOptions{Name: "SendEmail"})
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatal(err)
	}
}

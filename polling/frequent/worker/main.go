package main

import (
	"log"
	"time"

	"github.com/mfateev/samples-go-poc/polling"
	"github.com/mfateev/samples-go-poc/polling/frequent"

	isolateworker "github.com/mfateev/sdk-go-poc/worker"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
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

	w := isolateworker.New(c, frequent.TaskQueueName, worker.Options{})

	w.RegisterWorkflow(frequent.FrequentPolling)
	testService := polling.NewTestService(5)
	activities := &frequent.PollingActivities{
		TestService:  &testService,
		PollInterval: 1 * time.Second,
	}
	w.RegisterActivity(activities)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

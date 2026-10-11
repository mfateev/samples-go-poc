package main

import (
	"log"

	"github.com/mfateev/samples-go-poc/polling"
	"github.com/mfateev/samples-go-poc/polling/infrequent"

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

	w := isolateworker.New(c, infrequent.TaskQueueName, worker.Options{})

	w.RegisterWorkflow(infrequent.InfrequentPolling)
	testService := polling.NewTestService(5)
	activities := &infrequent.PollingActivities{
		TestService: &testService,
	}
	w.RegisterActivity(activities)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

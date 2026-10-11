package main

import (
	"context"
	"log"
	"time"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"

	"github.com/mfateev/samples-go-poc/polling/periodic_sequence"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "pollingSampleQueue_" + uuid.New(),
		TaskQueue: periodic_sequence.TaskQueueName,
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, "PeriodicSequencePolling", 1*time.Second)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}

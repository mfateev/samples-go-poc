package main

import (
	"context"
	"log"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
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

	fileID := uuid.New()
	workflowOptions := client.StartWorkflowOptions{
		ID:        "fileprocessing_" + fileID,
		TaskQueue: "fileprocessing",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, "SampleFileProcessingWorkflow", fileID)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}

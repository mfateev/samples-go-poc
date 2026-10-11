// @@@SNIPSTART samples-go-nexus-caller-starter
package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/mfateev/samples-go-poc/nexus/caller"
	"github.com/mfateev/samples-go-poc/nexus/options"
	"github.com/mfateev/samples-go-poc/nexus/service"
)

func main() {
	clientOptions, err := options.ParseClientOptionFlags(os.Args[1:])
	if err != nil {
		log.Fatalf("Invalid arguments: %v", err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()
	runWorkflow(c, "EchoCallerWorkflow", "Nexus Echo 👋")
	runWorkflow(c, "HelloCallerWorkflow", "Nexus", service.ES)
}

func runWorkflow(c client.Client, workflow string, args ...any) {
	ctx := context.Background()
	workflowOptions := client.StartWorkflowOptions{
		ID:        "nexus_hello_caller_workflow_" + time.Now().Format("20060102150405"),
		TaskQueue: caller.TaskQueue,
	}

	wr, err := c.ExecuteWorkflow(ctx, workflowOptions, workflow, args...)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", wr.GetID(), "RunID", wr.GetRunID())

	// Synchronously wait for the workflow completion.
	var result string
	err = wr.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Workflow result:", result)
}

// @@@SNIPEND

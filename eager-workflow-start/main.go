package main

import (
	"context"
	"log"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/worker"

	"github.com/mfateev/samples-go-poc/helloworld"
	isolateworker "github.com/mfateev/sdk-go-poc/worker"
)

const taskQueueName = "eager_wf_start"

func main() {
	// 1. Create the shared client.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// 2. Start the worker in a non-blocking manner before the workflow.
	workerOptions := worker.Options{
		OnFatalError: func(err error) { log.Fatalln("Worker error", err) },
	}
	w := isolateworker.New(c, taskQueueName, workerOptions)
	w.RegisterWorkflow(helloworld.HelloWorld)
	w.RegisterActivity(helloworld.Activity)
	err = w.Start()
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
	defer w.Stop()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "eager_wf" + uuid.New(),
		TaskQueue: taskQueueName,

		// 3. Set this flag to true to enable EWS.
		EnableEagerStart: true,
	}

	// 4. Reuse the client connection.
	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions,
		"HelloWorld", "Temporal")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	// 5. Wait for workflow completion.
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable to get workflow result", err)
	}
	log.Println("Workflow result:", result)
}

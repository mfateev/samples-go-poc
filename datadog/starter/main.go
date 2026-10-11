package main

import (
	"context"
	"log"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/datadog/tracing"
	"go.temporal.io/sdk/interceptor"
)

func main() {
	// Start the tracer and defer the Stop method.
	tracer.Start(tracer.WithAgentAddr("localhost:8126"))
	defer tracer.Stop()

	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		Interceptors: []interceptor.ClientInterceptor{tracing.NewTracingInterceptor(tracing.TracerOptions{})},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "datadog_workflow_id",
		TaskQueue: "datadog",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, "Workflow", "<param to log>")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	err = we.Get(context.Background(), nil)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Workflow completed. Check worker logs.")
}

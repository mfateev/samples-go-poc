package main

import (
	"context"
	"flag"
	"log"

	"go.temporal.io/sdk/client"

	"github.com/mfateev/samples-go-poc/pso"
)

func main() {
	var workflowID, runID, queryType string
	flag.StringVar(&workflowID, "w", "", "WorkflowID")
	flag.StringVar(&runID, "r", "", "RunID")
	flag.StringVar(&queryType, "t", "__stack_trace", "Query type is one of [__stack_trace, child, __open_sessions]")
	flag.Parse()

	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort:      client.DefaultHostPort,
		DataConverter: pso.NewJSONDataConverter(),
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	resp, err := c.QueryWorkflow(context.Background(), workflowID, runID, queryType)
	if err != nil {
		log.Fatalln("Unable to query workflow", err)
	}
	var result any
	if err := resp.Get(&result); err != nil {
		log.Fatalln("Unable to decode query result", err)
	}
	log.Println("Received query result", "Result", result)
}

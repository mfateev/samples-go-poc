package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	example "github.com/mfateev/samples-go-poc/random-streams"
	"go.temporal.io/sdk/client"
)

func main() {
	count := flag.Int("count", 10, "number of choices, 1..100")
	auditDraws := flag.Int("audit-draws", 0, "draws from the independent audit stream, 0..1000")
	flag.Parse()
	address := os.Getenv("TEMPORAL_ADDRESS")
	if address == "" {
		address = client.DefaultHostPort
	}
	c, err := client.Dial(client.Options{HostPort: address})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{ID: "random-streams-" + uuid.NewString(), TaskQueue: "random-streams-poc"}, "Plan", example.Input{Count: *count, AuditDraws: *auditDraws})
	if err != nil {
		log.Fatal(err)
	}
	var result example.Result
	if err := run.Get(ctx, &result); err != nil {
		log.Fatal(err)
	}
	log.Printf("Workflow %s: choices=%v jitter-ms=%v", run.GetID(), result.Choices, result.JitterMillis)
	for range 2 {
		query, err := c.QueryWorkflow(ctx, run.GetID(), run.GetRunID(), "next")
		if err != nil {
			log.Fatal(err)
		}
		var next int
		if err := query.Get(&next); err != nil {
			log.Fatal(err)
		}
		log.Printf("Next choice (query does not advance the workflow stream): %d", next)
	}
}

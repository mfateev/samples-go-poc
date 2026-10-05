package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.temporal.io/sdk/client"
)

func main() {
	address := os.Getenv("TEMPORAL_ADDRESS")
	if address == "" {
		address = client.DefaultHostPort
	}
	c, err := client.Dial(client.Options{HostPort: address})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	names := os.Args[1:]
	if len(names) == 0 {
		names = []string{"Ada", "Grace", "Linus"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID: fmt.Sprintf("goroutines-poc-%d", time.Now().UnixNano()), TaskQueue: "goroutines-poc",
	}, "GreetAll", names)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("workflow ID:", run.GetID())
	fmt.Println("run ID:", run.GetRunID())
	var greetings []string
	if err := run.Get(ctx, &greetings); err != nil {
		log.Fatal(err)
	}
	for _, greeting := range greetings {
		fmt.Println(greeting)
	}
}

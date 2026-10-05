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

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        fmt.Sprintf("sleep-for-days-poc-%d", time.Now().UnixNano()),
		TaskQueue: "sleep-for-days-poc",
	}, "SleepForDays")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("workflow ID:", run.GetID())
	fmt.Println("run ID:", run.GetRunID())
	fmt.Println("experimental: live execution requires concurrent bridge support")
	fmt.Println("send the complete signal to finish this workflow after that support is added")
}

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

	interval := (30 * 24 * time.Hour).String()
	if len(os.Args) > 1 {
		interval = os.Args[1]
	}
	if duration, err := time.ParseDuration(interval); err != nil || duration <= 0 {
		log.Fatal("interval must be a positive Go duration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        fmt.Sprintf("sleep-for-days-poc-%d", time.Now().UnixNano()),
		TaskQueue: "sleep-for-days-poc",
	}, "SleepForDays", []byte(interval))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("workflow ID:", run.GetID())
	fmt.Println("run ID:", run.GetRunID())
	fmt.Println("experimental: live execution requires concurrent bridge support")
	fmt.Println("send the complete signal to finish this workflow after that support is added")
}

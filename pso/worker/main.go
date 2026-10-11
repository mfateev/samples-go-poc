package main

import (
	"log"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/mfateev/samples-go-poc/pso"
	isolateworker "github.com/mfateev/sdk-go-poc/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort:      client.DefaultHostPort,
		DataConverter: pso.NewJSONDataConverter(),
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := isolateworker.New(c, "pso", worker.Options{
		MaxConcurrentActivityExecutionSize: 1, // Activities are supposed to be CPU intensive, so better limit the concurrency
	})

	if err := isolateworker.SetIsolateDataConverter(w, isolateworker.DataConverterOptions{Factory: pso.NewIsolateDataConverter}); err != nil {
		log.Fatal(err)
	}
	w.RegisterWorkflow(pso.PSOWorkflow)
	w.RegisterWorkflow(pso.PSOChildWorkflow)

	w.RegisterActivityWithOptions(pso.InitParticleActivity, activity.RegisterOptions{Name: pso.InitParticleActivityName})
	w.RegisterActivityWithOptions(pso.UpdateParticleActivity, activity.RegisterOptions{Name: pso.UpdateParticleActivityName})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

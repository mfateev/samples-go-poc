package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/worker"

	synchronousproxy "github.com/mfateev/samples-go-poc/synchronous-proxy"
	isolateworker "github.com/mfateev/sdk-go-poc/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := isolateworker.New(c, "ui-driven", worker.Options{})

	w.RegisterWorkflow(synchronousproxy.OrderWorkflow)
	w.RegisterWorkflow(synchronousproxy.UpdateOrderWorkflow)
	w.RegisterWorkflow(synchronousproxy.ShippingWorkflow)
	w.RegisterActivity(synchronousproxy.RegisterEmail)
	w.RegisterActivity(synchronousproxy.ValidateSize)
	w.RegisterActivity(synchronousproxy.ValidateColor)
	w.RegisterActivity(synchronousproxy.ScheduleDelivery)
	w.RegisterActivity(synchronousproxy.SendDeliveryEmail)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

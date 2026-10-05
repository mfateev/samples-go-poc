package main

import (
	"log"
	"os"

	choice "github.com/mfateev/samples-go-poc/choice-exclusive"
	"github.com/mfateev/sdk-go-poc/worker"
	"go.temporal.io/sdk/activity"
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
	w := worker.New(c, "choice-exclusive-poc", worker.Options{})
	w.RegisterWorkflow(choice.ExclusiveChoice)
	w.RegisterActivityWithOptions(choice.GetOrder, activity.RegisterOptions{Name: "GetOrder"})
	w.RegisterActivityWithOptions(choice.OrderApple, activity.RegisterOptions{Name: "OrderApple"})
	w.RegisterActivityWithOptions(choice.OrderBanana, activity.RegisterOptions{Name: "OrderBanana"})
	w.RegisterActivityWithOptions(choice.OrderCherry, activity.RegisterOptions{Name: "OrderCherry"})
	w.RegisterActivityWithOptions(choice.OrderOrange, activity.RegisterOptions{Name: "OrderOrange"})
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatal(err)
	}
}

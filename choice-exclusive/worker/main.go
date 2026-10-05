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
	orders := &choice.OrderActivities{OrderChoices: []string{
		choice.OrderChoiceApple, choice.OrderChoiceBanana,
		choice.OrderChoiceCherry, choice.OrderChoiceOrange,
	}}
	w.RegisterActivityWithOptions(orders.GetOrder, activity.RegisterOptions{Name: "GetOrder"})
	w.RegisterActivityWithOptions(orders.OrderApple, activity.RegisterOptions{Name: "OrderApple"})
	w.RegisterActivityWithOptions(orders.OrderBanana, activity.RegisterOptions{Name: "OrderBanana"})
	w.RegisterActivityWithOptions(orders.OrderCherry, activity.RegisterOptions{Name: "OrderCherry"})
	w.RegisterActivityWithOptions(orders.OrderOrange, activity.RegisterOptions{Name: "OrderOrange"})
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatal(err)
	}
}

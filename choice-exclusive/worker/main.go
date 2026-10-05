package main

import (
	"context"
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
	// The POC activity boundary carries bytes. Keep the sample's native
	// activity signatures and adapt them only when registering with the host.
	w.RegisterActivityWithOptions(func(_ context.Context, _ []byte) ([]byte, error) {
		selected, err := orders.GetOrder()
		return []byte(selected), err
	}, activity.RegisterOptions{Name: "GetOrder"})
	registerOrder := func(name string, order func(string) error) {
		w.RegisterActivityWithOptions(func(_ context.Context, input []byte) ([]byte, error) {
			return nil, order(string(input))
		}, activity.RegisterOptions{Name: name})
	}
	registerOrder("OrderApple", orders.OrderApple)
	registerOrder("OrderBanana", orders.OrderBanana)
	registerOrder("OrderCherry", orders.OrderCherry)
	registerOrder("OrderOrange", orders.OrderOrange)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatal(err)
	}
}

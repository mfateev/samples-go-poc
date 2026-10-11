package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	choice_multi "github.com/mfateev/samples-go-poc/choice-multi"
	isolateworker "github.com/mfateev/sdk-go-poc/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := isolateworker.New(c, "choice-multi", worker.Options{})

	w.RegisterWorkflow(choice_multi.MultiChoiceWorkflow)

	orderChoices := []string{
		choice_multi.OrderChoiceApple,
		choice_multi.OrderChoiceBanana,
		choice_multi.OrderChoiceCherry,
		choice_multi.OrderChoiceOrange}
	w.RegisterActivity(&choice_multi.OrderActivities{OrderChoices: orderChoices})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

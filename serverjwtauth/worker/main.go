package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/mfateev/samples-go-poc/helloworld"
	"github.com/mfateev/samples-go-poc/serverjwtauth"
	isolateworker "github.com/mfateev/sdk-go-poc/worker"
)

func main() {
	key, jwk, err := serverjwtauth.ReadKey()
	if err != nil {
		log.Fatalln(err)
	}
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		HeadersProvider: &serverjwtauth.JWTHeadersProvider{
			Config: serverjwtauth.JWTConfig{
				Key:   key,
				KeyID: jwk.KeyID,
				Permissions: []string{
					"default:read",
					"default:write",
				},
			},
		},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := isolateworker.New(c, "server-jwt-auth", worker.Options{})

	w.RegisterWorkflow(helloworld.HelloWorld)
	w.RegisterActivity(helloworld.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

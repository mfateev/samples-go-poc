package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mfateev/samples-go-poc/tracing/native"
	export "github.com/mfateev/sdk-go-poc/contrib/opentelemetry/exporter"

	clientupdate "github.com/mfateev/samples-go-poc/opentelemetry-v2/client-update-propagation"
	otelsetup "github.com/mfateev/samples-go-poc/opentelemetry-v2/setup"
	isolateworker "github.com/mfateev/sdk-go-poc/worker"
	"go.temporal.io/sdk/client"
	temporalotel "go.temporal.io/sdk/contrib/opentelemetry-v2"
	"go.temporal.io/sdk/worker"
)

const (
	serviceName     = "temporal-otel-v2-custom-client-update-propagation-worker"
	shutdownTimeout = 5 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	ctx := context.Background()
	tp, err := otelsetup.InitializeGlobalTracerProvider(ctx, serviceName)
	if err != nil {
		return fmt.Errorf("unable to create global tracer provider: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := tp.Shutdown(shutdownCtx); err != nil {
			log.Println("Error shutting down tracer provider:", err)
		}
	}()

	plugin, err := temporalotel.NewPlugin(temporalotel.PluginOptions{})
	if err != nil {
		return fmt.Errorf("unable to create plugin: %w", err)
	}
	c, err := client.Dial(client.Options{Plugins: []client.Plugin{plugin}})
	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}
	defer c.Close()

	w := isolateworker.New(c, clientupdate.TaskQueueName, worker.Options{})
	if err := isolateworker.SetIsolateInterceptors(w, isolateworker.InterceptorOptions{Factory: native.NewInterceptors, Config: []byte("v2-spans")}); err != nil {
		return err
	}
	if err := export.RegisterProcessor(w, native.SinkOp, otelsetup.Processor, nil); err != nil {
		return err
	}

	w.RegisterWorkflow(clientupdate.Workflow)

	if err := w.Run(worker.InterruptCh()); err != nil {
		return fmt.Errorf("worker run failed: %w", err)
	}
	return nil
}

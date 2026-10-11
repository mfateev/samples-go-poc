package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mfateev/samples-go-poc/tracing/native"
	export "github.com/mfateev/sdk-go-poc/contrib/opentelemetry/exporter"

	otelsetup "github.com/mfateev/samples-go-poc/opentelemetry-v2/setup"
	workflowactivity "github.com/mfateev/samples-go-poc/opentelemetry-v2/workflow-activity-propagation"
	isolateworker "github.com/mfateev/sdk-go-poc/worker"
	"go.temporal.io/sdk/client"
	temporalotel "go.temporal.io/sdk/contrib/opentelemetry-v2"
	"go.temporal.io/sdk/worker"
)

const (
	serviceName     = "temporal-otel-v2-custom-workflow-activity-propagation-worker"
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

	// @@@SNIPSTART samples-go-opentelemetry-v2-plugin-client
	plugin, err := temporalotel.NewPlugin(temporalotel.PluginOptions{})
	if err != nil {
		return fmt.Errorf("unable to create plugin: %w", err)
	}

	c, err := client.Dial(client.Options{Plugins: []client.Plugin{plugin}})
	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}
	defer c.Close()
	// @@@SNIPEND

	w := isolateworker.New(c, workflowactivity.TaskQueueName, worker.Options{})
	if err := isolateworker.SetIsolateInterceptors(w, isolateworker.InterceptorOptions{Factory: native.NewInterceptors, Config: []byte("v2")}); err != nil {
		return err
	}
	if err := export.RegisterProcessor(w, native.SinkOp, otelsetup.Processor, nil); err != nil {
		return err
	}

	w.RegisterWorkflow(workflowactivity.Workflow)
	w.RegisterActivity(workflowactivity.Activity)

	if err := w.Run(worker.InterruptCh()); err != nil {
		return fmt.Errorf("worker run failed: %w", err)
	}
	return nil
}

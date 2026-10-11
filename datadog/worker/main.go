package main

import (
	"io"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/mfateev/samples-go-poc/tracing/native"
	export "github.com/mfateev/sdk-go-poc/contrib/datadog/exporter"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/mfateev/samples-go-poc/datadog"
	isolateworker "github.com/mfateev/sdk-go-poc/worker"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/uber-go/tally/v4"
	"github.com/uber-go/tally/v4/prometheus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/datadog/tracing"
	sdktally "go.temporal.io/sdk/contrib/tally"
	"go.temporal.io/sdk/interceptor"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"
)

func main() {
	// Start the tracer and defer the Stop method.
	tracer.Start(tracer.WithAgentAddr("localhost:8126"))
	defer tracer.Stop()

	// Setup logging
	f, err := os.OpenFile("worker.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			log.Fatalf("error closing file: %v", err)
		}
	}()
	wrt := io.MultiWriter(os.Stdout, f)

	logger := tlog.NewStructuredLogger(
		slog.New(slog.NewJSONHandler(wrt, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})))

	c, err := client.Dial(client.Options{
		Logger:       logger,
		Interceptors: []interceptor.ClientInterceptor{tracing.NewTracingInterceptor(tracing.TracerOptions{})},
		MetricsHandler: sdktally.NewMetricsHandler(newPrometheusScope(prometheus.Configuration{
			ListenAddress: "localhost:9090",
			TimerType:     "histogram",
		})),
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := isolateworker.New(c, "datadog", worker.Options{})

	if err := isolateworker.SetIsolateInterceptors(w, isolateworker.InterceptorOptions{Factory: native.NewInterceptors, Config: []byte("datadog")}); err != nil {
		log.Fatal(err)
	}
	if err := export.Register(w, native.SinkOp, export.Options{Service: "temporal-datadog-sample"}); err != nil {
		log.Fatal(err)
	}
	w.RegisterWorkflow(datadog.Workflow)
	w.RegisterWorkflow(datadog.ChildWorkflow)
	w.RegisterActivity(datadog.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

func newPrometheusScope(c prometheus.Configuration) tally.Scope {
	reporter, err := c.NewReporter(
		prometheus.ConfigurationOptions{
			Registry: prom.NewRegistry(),
			OnError: func(err error) {
				log.Println("error in prometheus reporter", err)
			},
		},
	)
	if err != nil {
		log.Fatalln("error creating prometheus reporter", err)
	}
	scopeOpts := tally.ScopeOptions{
		CachedReporter:  reporter,
		Separator:       prometheus.DefaultSeparator,
		SanitizeOptions: &sdktally.PrometheusSanitizeOptions,
		Prefix:          "temporal_datadog",
	}
	scope, _ := tally.NewRootScope(scopeOpts, time.Second)
	scope = sdktally.NewPrometheusNamingScope(scope)

	log.Println("prometheus metrics scope created")
	return scope
}

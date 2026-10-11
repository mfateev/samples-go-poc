package main

import (
	"context"
	"log"
	"time"

	logger_interceptor "github.com/mfateev/samples-go-poc/logger-interceptor"
	isolateworker "github.com/mfateev/sdk-go-poc/worker"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	sdkinterceptor "go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := isolateworker.New(c, "logger-interceptor", worker.Options{
		// Create logger-interceptor that will put started time on the logger
		Interceptors: []sdkinterceptor.WorkerInterceptor{logger_interceptor.NewWorkerInterceptor(logger_interceptor.InterceptorOptions{
			GetExtraLogTagsForActivity: func(ctx context.Context) []any {
				return []any{"ActivityStartTime", activity.GetInfo(ctx).StartedTime.Format(time.RFC3339)}
			},
		})},
	})

	if err := isolateworker.SetIsolateInterceptors(w, isolateworker.InterceptorOptions{Factory: logger_interceptor.NewWorkflowInterceptors}); err != nil {
		log.Fatal(err)
	}
	w.RegisterWorkflow(logger_interceptor.Workflow)
	w.RegisterActivity(logger_interceptor.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

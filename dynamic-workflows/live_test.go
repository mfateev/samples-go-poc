package dynamic_workflows

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mfateev/sdk-go-poc/testsuite"
	isolateworker "github.com/mfateev/sdk-go-poc/worker"
	"github.com/stretchr/testify/require"
	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func TestDynamicWorkflowDevServerAndReplay(t *testing.T) {
	if os.Getenv("SAMPLES_DEV_SERVER") != "1" {
		t.Skip("set SAMPLES_DEV_SERVER=1 to start a Temporal development server")
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	server, err := testsuite.StartDevServer(ctx, testsuite.DevServerOptions{})
	require.NoError(t, err)
	defer func() { require.NoError(t, server.Stop()) }()
	c := server.Client()
	queue := fmt.Sprintf("dynamic-workflows-%d", time.Now().UnixNano())
	w := isolateworker.New(c, queue, worker.Options{})
	w.RegisterDynamicWorkflow(DynamicWorkflow, workflow.DynamicRegisterOptions{})
	w.RegisterDynamicActivity(DynamicActivity, activity.DynamicRegisterOptions{})
	require.NoError(t, w.Start())
	defer w.Stop()
	for _, example := range []struct{ name, first, second string }{
		{"DurableExecution", "Hello", "Temporal"},
		{"dynamic-activity", "Peanut", "Butter"},
		{"dynamic-activity123", "Jelly", "Time"},
	} {
		t.Run(example.name, func(t *testing.T) {
			run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{ID: queue + "-" + example.name, TaskQueue: queue}, example.name, example.first, example.second)
			require.NoError(t, err)
			var result string
			require.NoError(t, run.Get(ctx, &result))
			require.Equal(t, example.name+" - "+example.first+" - "+example.second, result)
			iterator := c.GetWorkflowHistory(ctx, run.GetID(), run.GetRunID(), false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
			history := new(historypb.History)
			for iterator.HasNext() {
				event, err := iterator.Next()
				require.NoError(t, err)
				history.Events = append(history.Events, event)
			}
			replayer := isolateworker.NewWorkflowReplayer()
			replayer.RegisterDynamicWorkflow(DynamicWorkflow, workflow.DynamicRegisterOptions{})
			require.NoError(t, replayer.ReplayWorkflowHistory(nil, history))
		})
	}
}

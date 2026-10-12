package random_streams

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
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/worker"
)

func TestNamedStreamsLiveAndReplay(t *testing.T) {
	if os.Getenv("SAMPLES_DEV_SERVER") != "1" {
		t.Skip("set SAMPLES_DEV_SERVER=1 to start a Temporal development server")
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	server, err := testsuite.StartDevServer(ctx, testsuite.DevServerOptions{})
	require.NoError(t, err)
	defer func() { require.NoError(t, server.Stop()) }()
	c := server.Client()
	defer worker.SetStickyWorkflowCacheSize(1000)
	for _, cache := range []int{1000, 0} {
		worker.SetStickyWorkflowCacheSize(cache)
		queue := fmt.Sprintf("random-streams-%d-%d", cache, time.Now().UnixNano())
		w := isolateworker.New(c, queue, worker.Options{})
		w.RegisterWorkflow(Plan)
		require.NoError(t, w.Start())
		func() {
			defer w.Stop()
			run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{ID: queue, TaskQueue: queue}, "Plan", Input{Count: 10, AuditDraws: 123})
			require.NoError(t, err)
			var result Result
			require.NoError(t, run.Get(ctx, &result))
			var first int
			for i := range 3 {
				query, err := c.QueryWorkflow(ctx, run.GetID(), run.GetRunID(), "next")
				require.NoError(t, err)
				var next int
				require.NoError(t, query.Get(&next))
				if i == 0 {
					first = next
				} else {
					require.Equal(t, first, next)
				}
			}
			iterator := c.GetWorkflowHistory(ctx, run.GetID(), run.GetRunID(), false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
			history := new(historypb.History)
			for iterator.HasNext() {
				event, err := iterator.Next()
				require.NoError(t, err)
				history.Events = append(history.Events, event)
			}
			require.Len(t, history.Events, 5, "named random draws must not create history events")
			completion := history.Events[len(history.Events)-1].GetWorkflowExecutionCompletedEventAttributes()
			require.NotNil(t, completion)
			var recorded Result
			require.NoError(t, converter.GetDefaultDataConverter().FromPayloads(completion.Result, &recorded))
			require.Equal(t, result, recorded)
			replayer := isolateworker.NewWorkflowReplayer()
			replayer.RegisterWorkflow(Plan)
			require.NoError(t, replayer.ReplayWorkflowHistory(nil, history))
			getter := replayer.(interface{ GetWorkflowResult(string, any) error })
			var replayed Result
			require.NoError(t, getter.GetWorkflowResult("", &replayed))
			require.Equal(t, recorded, replayed, "named stream output changed on replay")
		}()
	}
}

package reqrespactivity_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/mfateev/samples-go-poc/reqrespactivity"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	sdkwf "go.temporal.io/sdk/workflow"
)

func TestSignalBatchAndCallbacksDrainedBeforeContinueAsNew(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterActivity(reqrespactivity.UppercaseActivity)
	responses := make(map[string]string)
	var mu sync.Mutex // These are host activities, which really do execute concurrently.
	env.RegisterActivityWithOptions(func(_ context.Context, response *reqrespactivity.Response) error {
		mu.Lock()
		defer mu.Unlock()
		responses[response.ID] = response.Output
		return nil
	}, activity.RegisterOptions{Name: "reply"})
	env.RegisterDelayedCallback(func() {
		for index := range 550 {
			env.SignalWorkflow("request", &reqrespactivity.Request{
				ID: fmt.Sprint(index), Input: fmt.Sprintf("foo %d", index),
				ResponseActivity: "reply", ResponseTaskQueue: "replies",
			})
		}
	}, 0)
	env.ExecuteWorkflow(reqrespactivity.UppercaseWorkflow)
	var next *sdkwf.ContinueAsNewError
	require.True(t, errors.As(env.GetWorkflowError(), &next))
	mu.Lock()
	defer mu.Unlock()
	require.Len(t, responses, 550)
	for index := range 550 {
		require.Equal(t, fmt.Sprintf("FOO %d", index), responses[fmt.Sprint(index)])
	}
}

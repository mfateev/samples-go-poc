package reqrespactivity_test

import (
	"context"
	"testing"

	"github.com/mfateev/samples-go-poc/reqrespactivity"
	"github.com/mfateev/sdk-go-poc/testsuite"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
)

func TestUppercaseWorkflow(t *testing.T) {
	// Create env
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment(t)
	env.RegisterWorkflow(reqrespactivity.UppercaseWorkflow)
	env.RegisterActivity(reqrespactivity.UppercaseActivity)

	// Add another delayed callback to send a signal
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("request", &reqrespactivity.Request{
			ID:                "request1",
			Input:             "foo",
			ResponseActivity:  "external-activity",
			ResponseTaskQueue: "external-task-queue",
		})
	}, 2)

	// Add the external activity
	var externalResponses []*reqrespactivity.Response
	env.RegisterActivityWithOptions(
		func(_ context.Context, resp *reqrespactivity.Response) error {
			externalResponses = append(externalResponses, resp)
			return nil
		},
		activity.RegisterOptions{Name: "external-activity"},
	)

	// Run workflow
	env.ExecuteWorkflow(reqrespactivity.UppercaseWorkflow)

	// Confirm activity sent
	require.Equal(t, []*reqrespactivity.Response{{ID: "request1", Output: "FOO"}}, externalResponses)
}

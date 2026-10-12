package reqrespquery_test

import (
	"testing"
	"time"

	"github.com/mfateev/samples-go-poc/reqrespquery"
	"github.com/mfateev/sdk-go-poc/testsuite"
	"github.com/stretchr/testify/require"
)

func TestUppercaseWorkflow(t *testing.T) {
	// Create env
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment(t)
	env.RegisterWorkflow(reqrespquery.UppercaseWorkflow)
	env.RegisterActivity(reqrespquery.UppercaseActivity)

	// Add a delayed callback to send a signal
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("request", &reqrespquery.Request{ID: "request1", Input: "foo"})
	}, 1)

	// Add a delayed callback after that to query
	var queryResponse *reqrespquery.Response
	var queryErr error
	env.RegisterDelayedCallback(func() {
		val, err := env.QueryWorkflow("response", "request1")
		if err != nil {
			queryErr = err
		} else {
			queryErr = val.Get(&queryResponse)
		}
	}, 100*time.Millisecond)

	// Run workflow
	env.ExecuteWorkflow(reqrespquery.UppercaseWorkflow)

	// Confirm query response
	require.NoError(t, queryErr)
	require.Equal(t, &reqrespquery.Response{ID: "request1", Output: "FOO"}, queryResponse)
}

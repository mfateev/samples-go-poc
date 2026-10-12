package reqrespquery_test

import (
	"errors"
	"fmt"
	"github.com/mfateev/samples-go-poc/reqrespquery"
	"github.com/mfateev/sdk-go-poc/testsuite"
	"github.com/stretchr/testify/require"
	sdkwf "go.temporal.io/sdk/workflow"
	"testing"
)

func TestSignalBatchDrainedBeforeContinueAsNew(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment(t)
	env.RegisterActivity(reqrespquery.UppercaseActivity)
	env.RegisterDelayedCallback(func() {
		for index := range 550 {
			env.SignalWorkflow("request", &reqrespquery.Request{ID: fmt.Sprint(index), Input: fmt.Sprintf("foo %d", index)})
		}
	}, 0)
	env.ExecuteWorkflow(reqrespquery.UppercaseWorkflow)
	var next *sdkwf.ContinueAsNewError
	require.True(t, errors.As(env.GetWorkflowError(), &next))
	for index := range 550 {
		value, err := env.QueryWorkflow("response", fmt.Sprint(index))
		require.NoError(t, err)
		var response *reqrespquery.Response
		require.NoError(t, value.Get(&response))
		require.Equal(t, &reqrespquery.Response{ID: fmt.Sprint(index), Output: fmt.Sprintf("FOO %d", index)}, response)
	}
}

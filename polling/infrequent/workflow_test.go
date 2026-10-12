package infrequent

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mfateev/samples-go-poc/polling"

	"github.com/mfateev/sdk-go-poc/testsuite"
)

func Test_InfrequentPollingWorkflow(t *testing.T) {
	s := testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment(t)
	testService := polling.NewTestService(5)
	a := &PollingActivities{
		TestService: &testService,
	}
	env.RegisterActivity(a)

	env.ExecuteWorkflow(InfrequentPolling)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var pollResult string
	require.NoError(t, env.GetWorkflowResult(&pollResult))
	require.Equalf(t, pollResult, "OK", "The polling has returned the wrong result")
	env.AssertExpectations(t)
}

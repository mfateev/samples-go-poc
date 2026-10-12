package dynamic_workflows

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/mfateev/sdk-go-poc/testsuite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

func TestDynamicWorkflow(t *testing.T) {
	s := testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment(t)
	env.RegisterDynamicWorkflow(DynamicWorkflow, workflow.DynamicRegisterOptions{})
	env.RegisterDynamicActivity(DynamicActivity, activity.DynamicRegisterOptions{})
	env.ExecuteWorkflow("dynamic-activity", "Hello", "World")
	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())
	var result string
	err := env.GetWorkflowResult(&result)
	assert.NoError(t, err)
	assert.Equal(t, "dynamic-activity - Hello - World", result)
}

func TestDynamicWorkflowNamesAndArguments(t *testing.T) {
	for _, procs := range []int{1, 2, 8} {
		t.Run(fmt.Sprint(procs), func(t *testing.T) {
			previous := runtime.GOMAXPROCS(procs)
			defer runtime.GOMAXPROCS(previous)
			for _, name := range []string{"DurableExecution", "dynamic-activity", "dynamic-activity123"} {
				t.Run(name, func(t *testing.T) {
					var suite testsuite.WorkflowTestSuite
					env := suite.NewTestWorkflowEnvironment(t)
					env.RegisterDynamicWorkflow(DynamicWorkflow, workflow.DynamicRegisterOptions{})
					env.RegisterDynamicActivity(DynamicActivity, activity.DynamicRegisterOptions{})
					env.ExecuteWorkflow(name, "Hello", "World")
					require.True(t, env.IsWorkflowCompleted())
					require.NoError(t, env.GetWorkflowError())
					var result string
					require.NoError(t, env.GetWorkflowResult(&result))
					require.Equal(t, name+" - Hello - World", result)
				})
			}
		})
	}
}

func TestDynamicWorkflowInvalidArguments(t *testing.T) {
	for _, args := range [][]any{nil, {123, "wrong type"}} {
		var suite testsuite.WorkflowTestSuite
		env := suite.NewTestWorkflowEnvironment(t)
		env.RegisterDynamicWorkflow(DynamicWorkflow, workflow.DynamicRegisterOptions{})
		env.ExecuteWorkflow("arbitrary-workflow", args...)
		require.True(t, env.IsWorkflowCompleted())
		require.ErrorContains(t, env.GetWorkflowError(), "failed to decode arguments")
	}
}

func TestDynamicWorkflowPartialArguments(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment(t)
	env.RegisterDynamicWorkflow(DynamicWorkflow, workflow.DynamicRegisterOptions{})
	env.ExecuteWorkflow("arbitrary-workflow", "first")
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "arbitrary-workflow - first - ", result)
}

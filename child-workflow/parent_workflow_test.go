package child_workflow_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/mfateev/sdk-go-poc/testsuite"
	"github.com/stretchr/testify/require"

	"github.com/mfateev/samples-go-poc/child-workflow"
)

func Test_Workflow_Integration(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment(t)
	env.RegisterWorkflow(child_workflow.SampleChildWorkflow)

	env.ExecuteWorkflow(child_workflow.SampleParentWorkflow)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "Hello World!", result)
}

func Test_Workflow_Isolation(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment(t)
	env.RegisterWorkflow(child_workflow.SampleChildWorkflow)
	env.OnWorkflow("SampleChildWorkflow", mock.Anything, mock.Anything).Return(
		func(ctx context.Context, name string) (string, error) {
			return "Hi " + name + "!", nil
		},
	)

	env.ExecuteWorkflow(child_workflow.SampleParentWorkflow)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "Hi World!", result)
}

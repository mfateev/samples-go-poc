package memo

import (
	"testing"

	"github.com/mfateev/sdk-go-poc/testsuite"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment(t)
	env.RegisterWorkflow(MemoWorkflow)
	env.RegisterActivity(DescribeWorkflow)

	// mock search attributes on start
	_ = env.SetMemoOnStart(map[string]any{"description": "Test memo workflow"})

	// mock upsert operations
	memo := map[string]any{
		"description": "Test upsert memo workflow",
	}
	env.OnUpsertMemo(mock.MatchedBy(func(values map[string]any) bool {
		if len(values) != len(memo) {
			return false
		}
		for key, want := range memo {
			raw, ok := values[key].(converter.RawValue)
			if !ok {
				return false
			}
			var got string
			if err := converter.GetDefaultDataConverter().FromPayload(raw.Payload(), &got); err != nil || got != want {
				return false
			}
		}
		return true
	})).Return(nil).Once()

	// mock activity
	env.OnActivity(DescribeWorkflow, mock.Anything, mock.Anything).Return(map[string]any{}, nil).Once()

	env.ExecuteWorkflow(MemoWorkflow)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}

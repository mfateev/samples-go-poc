package batch_sliding_window

import (
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	sdkwf "go.temporal.io/sdk/workflow"
	"testing"
)

func TestDrainBeforeContinueAsNew(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterActivity(&RecordLoader{RecordCount: 20})
	env.OnActivity((&RecordLoader{}).GetRecords, mock.Anything, mock.Anything).Return(GetRecordsOutput{}, nil)
	// All records here were started by the previous run. A backlog admitted
	// before channel registration must be drained into the continuation input.
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("ReportCompletion", 1)
		env.SignalWorkflow("ReportCompletion", 2)
		env.SignalWorkflow("ReportCompletion", 1) // Retry must not count twice.
	}, 0)
	env.ExecuteWorkflow(SlidingWindowWorkflow, SlidingWindowWorkflowInput{
		PageSize: 2, SlidingWindowSize: 3, Offset: 10, MaximumOffset: 20,
		CurrentRecords: map[int]bool{1: true, 2: true}, Progress: 7,
	})
	// A zero-record page still demonstrates continuation of the existing state.
	var next *sdkwf.ContinueAsNewError
	require.True(t, errors.As(env.GetWorkflowError(), &next))
	var input SlidingWindowWorkflowInput
	require.NoError(t, converter.GetDefaultDataConverter().FromPayloads(next.Input, &input))
	require.Equal(t, 9, input.Progress)
	require.Empty(t, input.CurrentRecords)
}

func TestFinalRunWaitsForAllCompletions(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("ReportCompletion", 1)
		env.SignalWorkflow("ReportCompletion", 1)
		env.SignalWorkflow("ReportCompletion", 2)
	}, 0)
	env.ExecuteWorkflow(SlidingWindowWorkflow, SlidingWindowWorkflowInput{
		PageSize: 2, SlidingWindowSize: 2, Offset: 10, MaximumOffset: 10,
		CurrentRecords: map[int]bool{1: true, 2: true}, Progress: 7,
	})
	require.NoError(t, env.GetWorkflowError())
	var count int
	require.NoError(t, env.GetWorkflowResult(&count))
	require.Equal(t, 9, count)
}

func TestSlidingWindowWithNativeChildren(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(RecordProcessorWorkflow)
	env.RegisterActivity(&RecordLoader{RecordCount: 3})
	env.ExecuteWorkflow(SlidingWindowWorkflow, SlidingWindowWorkflowInput{
		PageSize: 4, SlidingWindowSize: 2, MaximumOffset: 3,
	})
	require.NoError(t, env.GetWorkflowError())
	var count int
	require.NoError(t, env.GetWorkflowResult(&count))
	require.Equal(t, 3, count)
}

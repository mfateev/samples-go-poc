package sleepfordays

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mfateev/sdk-go-poc/testsuite"
)

func TestSleepForDaysWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment(t)

	numActivityCalls := 0
	env.RegisterActivity(SendEmail)
	env.OnActivity(SendEmail, mock.Anything, mock.Anything).Run(
		func(args mock.Arguments) { numActivityCalls++ },
	).Return(nil)

	startTime := env.Now()

	// Time-skip 90 days.
	env.RegisterDelayedCallback(func() {
		// Check that the activity has been called 3 times.
		require.Equal(t, 3, numActivityCalls)
		// Send the signal to complete the workflow.
		env.SignalWorkflow("complete", nil)
		// Expect no more activity calls to have been made - workflow is complete.
		require.Equal(t, 3, numActivityCalls)
		// Expect more than 90 days to have passed.
		require.Equal(t, env.Now().Sub(startTime), time.Hour*24*90)
	}, time.Hour*24*90)

	// Execute workflow.
	env.ExecuteWorkflow(SleepForDays)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}

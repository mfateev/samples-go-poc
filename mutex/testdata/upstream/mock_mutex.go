package mutex

import (
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
)

// MockMutexLock stubs mutex.Lock call
func MockMutexLock(env *testsuite.TestWorkflowEnvironment, resourceID string, mockError error) {
	execution := &workflow.Execution{ID: "mockID", RunID: "mockRunID"}
	env.OnActivity(SignalWithStartMutexWorkflowActivity,
		mock.Anything, mock.Anything, resourceID, mock.Anything, mock.Anything).
		Return(execution, mockError)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(AcquireLockSignalName, "mockReleaseLockChannelName")
	}, time.Millisecond*0)
	if mockError == nil {
		env.OnSignalExternalWorkflow(mock.Anything, mock.Anything, execution.RunID,
			mock.Anything, mock.Anything).Return(nil)
	}
}

package saga

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mfateev/sdk-go-poc/testsuite"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment(t)
	// Mock activity implementation
	testDetails := TransferDetails{
		Amount:      1.00,
		FromAccount: "001-001",
		ToAccount:   "002-002",
		ReferenceID: "1234",
	}
	env.OnActivity(Withdraw, mock.Anything, testDetails).Return(nil)
	env.OnActivity(WithdrawCompensation, mock.Anything, testDetails).Return(nil)
	env.OnActivity(Deposit, mock.Anything, testDetails).Return(nil)
	env.OnActivity(DepositCompensation, mock.Anything, testDetails).Return(nil)
	env.OnActivity(StepWithError, mock.Anything, testDetails).Return(errors.New("some error"))
	env.ExecuteWorkflow(TransferMoney, testDetails)
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
}

package earlyreturn

import (
	"fmt"
	"testing"

	"github.com/mfateev/sdk-go-poc/testsuite"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_CompleteTransaction_Succeeds(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment(t)

	txRequest := TransactionRequest{SourceAccount: "Bob", TargetAccount: "Alice", Amount: 100}
	env.RegisterActivity(InitTransaction)
	env.RegisterActivity(CompleteTransaction)

	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(UpdateName, uuid.New(), &testsuite.TestUpdateCallback{
			OnAccept: func() {},
			OnReject: func(err error) {
				require.Fail(t, "unexpected rejection")
			},
			OnComplete: func(i any, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, testsuite.DecodeResult[*Transaction](i).ID)
			},
		})
	}, 0) // NOTE: zero delay ensures Update is delivered in first workflow task
	env.ExecuteWorkflow(Workflow, txRequest)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}

func Test_CompleteTransaction_Fails(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment(t)

	txRequest := TransactionRequest{SourceAccount: "Bob", TargetAccount: "Alice", Amount: 100}
	env.RegisterActivity(InitTransaction)
	env.RegisterActivity(CompleteTransaction)

	env.OnActivity(CompleteTransaction, mock.Anything, mock.Anything).Return(fmt.Errorf("crash"))

	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(UpdateName, uuid.New(), &testsuite.TestUpdateCallback{
			OnAccept: func() {},
			OnReject: func(err error) {
				require.Fail(t, "unexpected rejection")
			},
			OnComplete: func(i any, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, testsuite.DecodeResult[*Transaction](i).ID)
			},
		})
	}, 0)
	env.ExecuteWorkflow(Workflow, txRequest)

	require.True(t, env.IsWorkflowCompleted())
	require.ErrorContains(t, env.GetWorkflowError(), "crash")
}

func Test_CancelTransaction(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment(t)

	txRequest := TransactionRequest{SourceAccount: "Bob", TargetAccount: "Alice", Amount: -1} // invalid!
	env.RegisterActivity(InitTransaction)
	env.RegisterActivity(CancelTransaction)

	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(UpdateName, uuid.New(), &testsuite.TestUpdateCallback{
			OnAccept: func() {},
			OnReject: func(err error) {
				require.Fail(t, "unexpected rejection")
			},
			OnComplete: func(i any, err error) {
				require.ErrorContains(t, err, "invalid Amount")
			},
		})
	}, 0)
	env.ExecuteWorkflow(Workflow, txRequest)

	require.True(t, env.IsWorkflowCompleted())
	require.ErrorContains(t, env.GetWorkflowError(), "invalid Amount")
}

func Test_CancelTransaction_Fails(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment(t)

	txRequest := TransactionRequest{SourceAccount: "Bob", TargetAccount: "Alice", Amount: -1} // invalid!
	env.RegisterActivity(InitTransaction)
	env.RegisterActivity(CancelTransaction)

	env.OnActivity(CancelTransaction, mock.Anything, mock.Anything).Return(fmt.Errorf("crash"))

	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(UpdateName, uuid.New(), &testsuite.TestUpdateCallback{
			OnAccept: func() {},
			OnReject: func(err error) {
				require.Fail(t, "unexpected rejection")
			},
			OnComplete: func(i any, err error) {
				require.ErrorContains(t, err, "invalid Amount")
			},
		})
	}, 0)
	env.ExecuteWorkflow(Workflow, txRequest)

	require.True(t, env.IsWorkflowCompleted())
	require.ErrorContains(t, env.GetWorkflowError(), "crash")
}

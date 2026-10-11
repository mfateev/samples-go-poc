package shoppingcart

import (
	"testing"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"

	"go.temporal.io/sdk/testsuite"
)

func Test_ShoppingCartWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	updatesCompleted := 0

	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(UpdateName, uuid.New(), &testsuite.TestUpdateCallback{
			OnAccept: func() {},
			OnReject: func(err error) { require.Fail(t, "unexpected rejection") },
			OnComplete: func(i any, err error) {
				require.NoError(t, err)
				cartState := testsuite.DecodeResult[*CartState](i)
				ok := cartState != nil
				if !ok {
					require.Fail(t, "Invalid return type")
				}
				require.Equal(t, cartState.Items["apple"], 1)
				updatesCompleted++
			},
		}, "add", "apple")
	}, 0)

	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(UpdateName, uuid.New(), &testsuite.TestUpdateCallback{
			OnAccept: func() {},
			OnReject: func(err error) { require.Fail(t, "unexpected rejection") },
			OnComplete: func(i any, err error) {
				require.NoError(t, err)
				cartState := testsuite.DecodeResult[*CartState](i)
				ok := cartState != nil
				if !ok {
					require.Fail(t, "Invalid return type")
				}
				_, ok = cartState.Items["apple"]
				require.False(t, ok)
				updatesCompleted++
			},
		}, "remove", "apple")
	}, 0)

	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(UpdateName, uuid.New(), &testsuite.TestUpdateCallback{
			OnAccept: func() { require.Fail(t, "unexpected accept") },
			OnReject: func(err error) {
				require.Error(t, err)
				require.EqualError(t, err, "unsupported action type: invalid")
			},
			OnComplete: func(i any, err error) {
			},
		}, "invalid", "apple")
	}, 0)

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("checkout", nil)
	}, 0)
	env.ExecuteWorkflow(CartWorkflow, nil)

	require.True(t, env.IsWorkflowCompleted())
	require.Equal(t, 2, updatesCompleted)
}

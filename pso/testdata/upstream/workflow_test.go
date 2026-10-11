package pso

import (
	"context"
	"strings"
	"testing"

	isolateworker "github.com/mfateev/sdk-go-poc/worker"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(PSOChildWorkflow)

	env.RegisterActivityWithOptions(
		InitParticleActivity,
		activity.RegisterOptions{Name: InitParticleActivityName},
	)
	env.RegisterActivityWithOptions(
		UpdateParticleActivity,
		activity.RegisterOptions{Name: UpdateParticleActivityName},
	)

	var activityCalled []string

	var dataConverter = NewJSONDataConverter()
	env.SetDataConverter(dataConverter)
	require.NoError(t, env.SetIsolateDataConverter(isolateworker.DataConverterOptions{Factory: NewIsolateDataConverter}))

	// env.SetWorkflowTimeout(5 * time.Minute)
	// env.SetTestTimeout(5 * time.Minute)

	env.SetOnActivityStartedListener(func(activityInfo *activity.Info, ctx context.Context, args converter.EncodedValues) {
		activityType := activityInfo.ActivityType.Name
		activityCalled = append(activityCalled, activityType)
		switch activityType {
		case "initParticleActivityName":
		case "updateParticleActivityName":
		default:
			panic("unexpected activity call")
		}
	})

	// The native workflow derives its child ID from the parent and attempt.
	childWorkflowID := "default-test-workflow-id/pso/1"

	env.ExecuteWorkflow(PSOWorkflow, "sphere")

	require.True(t, env.IsWorkflowCompleted())
	queryAndVerify(t, env, "child", childWorkflowID)
	//queryAndVerify(t, env, "iteration", "???")
	// consider recreating a new test env on every iteration and calling execute workflow
	// with the arguments from the previous iteration (contained in ContinueAsNewError)
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.True(t, strings.HasPrefix(result, "Optimization was successful at attempt #1"))
}

func queryAndVerify(t *testing.T, env *testsuite.TestWorkflowEnvironment, query string, expectedState string) {
	result, err := env.QueryWorkflow(query)
	require.NoError(t, err)
	var state string
	err = result.Get(&state)
	require.NoError(t, err)
	require.Equal(t, expectedState, state)
}

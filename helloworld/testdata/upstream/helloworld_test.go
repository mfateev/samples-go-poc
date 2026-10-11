package helloworld

import (
	"context"
	"os"
	"testing"

	isolateworker "github.com/mfateev/sdk-go-poc/worker"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock activity implementation
	env.OnActivity(Activity, mock.Anything, "Temporal").Return("Hello Temporal!", nil)

	env.ExecuteWorkflow(HelloWorld, "Temporal")

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "Hello Temporal!", result)
}

func Test_Activity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(Activity)

	val, err := env.ExecuteActivity(Activity, "World")
	require.NoError(t, err)

	var res string
	require.NoError(t, val.Get(&res))
	require.Equal(t, "Hello World!", res)
}

func Test_Using_DevServer(t *testing.T) {
	if os.Getenv("SAMPLES_DEV_SERVER") != "1" {
		t.Skip("set SAMPLES_DEV_SERVER=1 to start a Temporal development server")
	}
	server, err := testsuite.StartDevServer(context.Background(), testsuite.DevServerOptions{})
	require.NoError(t, err)
	defer func() { require.NoError(t, server.Stop()) }()
	c := server.Client()
	w := isolateworker.New(c, "hello-world", worker.Options{})
	w.RegisterWorkflow(HelloWorld)
	w.RegisterActivity(Activity)
	require.NoError(t, w.Start())
	defer w.Stop()
	run, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{ID: "hello_world_workflowID", TaskQueue: "hello-world"}, "HelloWorld", "Temporal")
	require.NoError(t, err)
	var result string
	require.NoError(t, run.Get(context.Background(), &result))
	require.Equal(t, "Hello Temporal!", result)
}

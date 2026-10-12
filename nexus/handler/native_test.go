package handler

import (
	"testing"

	"github.com/mfateev/samples-go-poc/nexus/service"
	"github.com/mfateev/sdk-go-poc/testsuite"
	"github.com/stretchr/testify/require"
)

func TestNativeHandler(t *testing.T) {
	for _, test := range []struct {
		language service.Language
		want     string
	}{
		{service.EN, "Hello Ada 👋"},
		{service.FR, "Bonjour Ada 👋"},
		{service.DE, "Hallo Ada 👋"},
		{service.ES, "¡Hola! Ada 👋"},
		{service.TR, "Merhaba Ada 👋"},
	} {
		t.Run(string(test.language), func(t *testing.T) {
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment(t)
			env.ExecuteWorkflow(HelloHandlerWorkflow, service.HelloInput{Name: "Ada", Language: test.language})
			require.NoError(t, env.GetWorkflowError())
			var output service.HelloOutput
			require.NoError(t, env.GetWorkflowResult(&output))
			require.Equal(t, test.want, output.Message)
		})
	}
}

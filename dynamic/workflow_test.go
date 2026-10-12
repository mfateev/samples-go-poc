package dynamic

import (
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mfateev/sdk-go-poc/testsuite"
)

func TestDynamicWorkflow(t *testing.T) {
	s := testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment(t)
	env.RegisterActivity(&Activities{})
	var a *Activities
	env.OnActivity(a.GetGreeting, mock.Anything).Return("Greet", nil).Times(1)
	env.OnActivity(a.GetName, mock.Anything).Return("Name", nil).Times(1)
	env.OnActivity(a.SayGreeting, mock.Anything, "Greet", "Name").Return("Greet Name", nil).Times(1)

	env.ExecuteWorkflow(SampleGreetingsWorkflow)

	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}

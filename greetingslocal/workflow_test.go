package greetingslocal

import (
	"testing"

	"github.com/mfateev/sdk-go-poc/testsuite"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_SampleGreetingsWorkflow() {
	env := s.NewTestWorkflowEnvironment(s.T())
	var a *Activities
	env.RegisterActivity(a)
	//env.RegisterActivity(a)

	env.OnActivity(a.GetGreeting, mock.Anything).Return("Hello", nil)
	env.OnActivity(a.GetName, mock.Anything).Return("World", nil)
	env.OnActivity(a.SayGreeting, mock.Anything, "Hello", "World").Return("Hello World!", nil)

	env.ExecuteWorkflow(GreetingSample)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	env.AssertExpectations(s.T())
}

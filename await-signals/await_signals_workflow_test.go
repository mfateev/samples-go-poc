package await_signals

import (
	"testing"
	"time"

	"github.com/mfateev/sdk-go-poc/testsuite"
	"github.com/stretchr/testify/suite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_WorkflowTimeout() {
	env := s.NewTestWorkflowEnvironment(s.T())
	env.ExecuteWorkflow(AwaitSignalsWorkflow)

	s.True(env.IsWorkflowCompleted())
	// Workflow times out
	s.Error(env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_SignalsInOrder() {
	env := s.NewTestWorkflowEnvironment(s.T())
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("Signal1", nil)
	}, time.Hour)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("Signal2", nil)
	}, time.Hour+time.Second)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("Signal3", nil)
	}, time.Hour+3*time.Second)
	env.ExecuteWorkflow(AwaitSignalsWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_SignalsInReverseOrder() {
	env := s.NewTestWorkflowEnvironment(s.T())
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("Signal3", nil)
	}, time.Hour)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("Signal2", nil)
	}, time.Hour+time.Second)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("Signal1", nil)
	}, time.Hour+3*time.Second)
	env.ExecuteWorkflow(AwaitSignalsWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

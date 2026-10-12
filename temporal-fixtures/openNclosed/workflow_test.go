package openNclosed

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/worker"

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

func (s *UnitTestSuite) Test_LargePayloadWorkflow() {
	env := s.NewTestWorkflowEnvironment(s.T())
	env.SetWorkerOptions(worker.Options{})
	env.RegisterActivity(Activity)

	keepOpen := false
	env.OnActivity(Activity, mock.Anything, keepOpen).Return("hello", nil)

	env.ExecuteWorkflow(OpenNClosedWorkflow, keepOpen)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

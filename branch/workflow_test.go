// @@@SNIPSTART samples-go-branch-workflow-definition-test
package branch

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/mock"

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

func (s *UnitTestSuite) Test_BranchWorkflow() {
	env := s.NewTestWorkflowEnvironment(s.T())
	env.RegisterActivity(SampleActivity)
	env.OnActivity(SampleActivity, mock.Anything, mock.Anything).Return("one", nil).Once()
	env.OnActivity(SampleActivity, mock.Anything, mock.Anything).Return("two", nil).Once()
	env.OnActivity(SampleActivity, mock.Anything, mock.Anything).Return("three", nil).Once()
	env.ExecuteWorkflow(SampleBranchWorkflow, 3)
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var result []string
	s.NoError(env.GetWorkflowResult(&result))
	sort.Strings(result)
	expected := []string{"one", "two", "three"}
	sort.Strings(expected)
	s.Equal(expected, result)
}

// @@@SNIPEND

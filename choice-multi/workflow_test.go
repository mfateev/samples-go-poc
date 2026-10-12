package choice_multi

import (
	"testing"

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

func (s *UnitTestSuite) Test_MultiChoiceWorkflow() {
	env := s.NewTestWorkflowEnvironment(s.T())

	orderChoices := []string{
		OrderChoiceApple,
		OrderChoiceBanana,
		OrderChoiceOrange}
	env.RegisterActivity(&OrderActivities{OrderChoices: orderChoices})

	env.ExecuteWorkflow(MultiChoiceWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

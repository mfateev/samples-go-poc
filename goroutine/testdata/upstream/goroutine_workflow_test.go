package goroutine

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func Step1Test(ctx context.Context, input string) (output string, err error) {
	return input + ", Step1", nil
}

func Step2Test(ctx context.Context, input string) (output string, err error) {
	return input + ", Step2", nil
}

func (s *UnitTestSuite) Test_GoroutineWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	env.RegisterActivityWithOptions(Step1Test, activity.RegisterOptions{Name: "Step1"})
	env.RegisterActivityWithOptions(Step2Test, activity.RegisterOptions{Name: "Step2"})

	env.ExecuteWorkflow(SampleGoroutineWorkflow, 2)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	var result []string
	_ = env.GetWorkflowResult(&result)

	fmt.Println(result)
	sort.Strings(result) // Order of goroutine execution is not defined
	s.Equal("0, Step1, Step2", result[0])
	s.Equal("1, Step1, Step2", result[1])
}

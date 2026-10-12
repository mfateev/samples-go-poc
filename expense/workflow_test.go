package expense

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func (s *UnitTestSuite) Test_WorkflowWithMockActivities() {
	env := s.NewTestWorkflowEnvironment(s.T())
	env.RegisterActivity(CreateExpenseActivity)
	env.RegisterActivity(WaitForDecisionActivity)
	env.RegisterActivity(PaymentActivity)

	env.OnActivity(CreateExpenseActivity, mock.Anything, mock.Anything).Return(nil).Once()
	env.OnActivity(WaitForDecisionActivity, mock.Anything, mock.Anything).Return("APPROVED", nil).Once()
	env.OnActivity(PaymentActivity, mock.Anything, mock.Anything).Return(nil).Once()

	env.ExecuteWorkflow(SampleExpenseWorkflow, "test-expense-id")

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var workflowResult string
	err := env.GetWorkflowResult(&workflowResult)
	s.NoError(err)
	s.Equal("COMPLETED", workflowResult)
	env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_WorkflowWithMockServer() {
	env := s.NewTestWorkflowEnvironment(s.T())
	env.RegisterActivity(CreateExpenseActivity)
	env.RegisterActivity(WaitForDecisionActivity)
	env.RegisterActivity(PaymentActivity)

	// setup mock expense server
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/text")
		switch r.URL.Path {
		case "/create":
		case "/registerCallback":
			taskToken := []byte(r.PostFormValue("task_token"))
			// simulate the expense is approved one hour later.
			env.RegisterDelayedCallback(func() {
				_ = env.CompleteActivity(taskToken, "APPROVED", nil)
			}, time.Hour)
		case "/action":
		}
		_, _ = io.WriteString(w, "SUCCEED")
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	// pointing server to test mock
	ExpenseServerHostPort = server.URL

	env.ExecuteWorkflow(SampleExpenseWorkflow, "test-expense-id")

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var workflowResult string
	err := env.GetWorkflowResult(&workflowResult)
	s.NoError(err)
	s.Equal("COMPLETED", workflowResult)
	env.AssertExpectations(s.T())
}

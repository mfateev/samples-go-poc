package fileprocessing

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

func (s *UnitTestSuite) Test_SampleFileProcessingWorkflow() {
	env := s.NewTestWorkflowEnvironment(s.T())
	env.SetWorkerOptions(worker.Options{
		EnableSessionWorker: true, // Important for a worker to participate in the session
	})
	var a *Activities
	env.RegisterActivity(a)

	env.OnActivity(a.DownloadFileActivity, mock.Anything, "file1").Return("file2", nil)
	env.OnActivity(a.ProcessFileActivity, mock.Anything, "file2").Return("file3", nil)
	env.OnActivity(a.UploadFileActivity, mock.Anything, "file3").Return(nil)

	env.ExecuteWorkflow(SampleFileProcessingWorkflow, "file1")

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	env.AssertExpectations(s.T())
}

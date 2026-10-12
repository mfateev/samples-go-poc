package helloworld

import (
	"testing"

	"github.com/golang/mock/gomock"
	isolateworker "github.com/mfateev/sdk-go-poc/worker"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/api/workflowservicemock/v1"
	"go.temporal.io/sdk/workflow"
)

type replayTestSuite struct {
	suite.Suite
	mockCtrl *gomock.Controller
	service  *workflowservicemock.MockWorkflowServiceClient
}

func TestReplayTestSuite(t *testing.T) {
	s := new(replayTestSuite)
	suite.Run(t, s)
}

func (s *replayTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.service = workflowservicemock.NewMockWorkflowServiceClient(s.mockCtrl)
}

func (s *replayTestSuite) TearDownTest() {
	s.mockCtrl.Finish() // assert mock’s expectations
}

// This replay test is the recommended way to make sure changing workflow code is backward compatible without non-deterministic errors.
// "helloworld.json" can be downloaded from Temporal CLI:
//
//	temporal workflow show --workflow-id hello_world_workflowID --outputjson > helloworld.json
//
// Or from Temporal Web UI. And you may need to change workflowType in the first event.
func (s *replayTestSuite) TestReplayWorkflowHistoryFromFile() {
	replayer := isolateworker.NewWorkflowReplayer()

	replayer.RegisterWorkflowWithOptions(HelloWorld, workflow.RegisterOptions{Name: "Workflow"})

	err := replayer.ReplayWorkflowHistoryFromJSONFile(nil, "helloworld.json")
	require.NoError(s.T(), err)
}

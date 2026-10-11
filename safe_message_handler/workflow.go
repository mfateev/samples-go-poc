package safe_message_handler

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"context"

	"github.com/mfateev/sdk-go-poc/workflow"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
)

const (
	StartCluster      = "StartCluster"
	ShutdownCluster   = "ShutdownCluster"
	AssignNodesToJobs = "AssignNodesToJobs"
	DeleteJob         = "DeleteJob"
)

type (
	// In workflows that continue-as-new, it's convenient to store all your state in one serializable structure
	// to make it easier to pass between runs
	ClusterManagerState struct {
		ClusterStarted  bool
		ClusterShutdown bool
		Nodes           map[string]string
		JobsAssigned    map[string]struct{}
	}

	ClusterManagerInput struct {
		State             *ClusterManagerState
		TestContinueAsNew bool
	}

	ClusterManagerResult struct {
		NumCurrentlyAssignedNodes int
		NumBadNodes               int
	}

	// Be in the habit of storing message inputs and outputs in serializable structures.
	// This makes it easier to add more over time in a backward-compatible way.
	ClusterManagerAssignNodesToJobInput struct {
		// If larger or smaller than previous amounts, will resize the job.
		TotalNumNodes int
		JobName       string
	}

	ClusterManagerDeleteJobInput struct {
		JobName string
	}

	ClusterManagerAssignNodesToJobResult struct {
		NodesAssigned map[string]struct{}
	}

	ClusterManager struct {
		state             ClusterManagerState
		nodeLock          *sync.Mutex
		logger            log.Logger
		sleepInterval     time.Duration
		maxHistoryLength  int
		startCh           <-chan any
		started           chan struct{}
		shutdownCh        <-chan any
		testContinueAsNew bool
	}
)

func newClusterManager(ctx context.Context, wfInput ClusterManagerInput) (*ClusterManager, error) {
	logger := workflow.GetLogger(ctx)
	sleepInterval := time.Second * 600
	maxHistoryLength := 0
	nodeLock := &sync.Mutex{}

	state := ClusterManagerState{
		Nodes:        make(map[string]string),
		JobsAssigned: make(map[string]struct{}),
	}
	if wfInput.State != nil {
		state = *wfInput.State
	}
	if wfInput.TestContinueAsNew {
		maxHistoryLength = 120
		sleepInterval = time.Second * 1
	}

	startCh := workflow.GetSignalChannel[any](ctx, StartCluster)
	shutdownCh := workflow.GetSignalChannel[any](ctx, ShutdownCluster)

	state.Nodes = make(map[string]string)
	for i := range 25 {
		state.Nodes[fmt.Sprint(i)] = ""
	}

	cm := &ClusterManager{
		state:             state,
		started:           make(chan struct{}),
		nodeLock:          nodeLock,
		logger:            logger,
		startCh:           startCh,
		shutdownCh:        shutdownCh,
		sleepInterval:     sleepInterval,
		maxHistoryLength:  maxHistoryLength,
		testContinueAsNew: wfInput.TestContinueAsNew,
	}

	err := workflow.SetUpdateHandler(ctx, AssignNodesToJobs, cm.AssignNodesToJobs)
	if err != nil {
		return nil, err
	}

	err = workflow.SetUpdateHandler(ctx, DeleteJob, cm.DeleteJob)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func (cm *ClusterManager) badNodes() map[string]struct{} {
	badNodes := make(map[string]struct{})
	for k := range cm.state.Nodes {
		if cm.state.Nodes[k] == "BAD!" {
			badNodes[k] = struct{}{}
		}
	}
	return badNodes
}

// This is an update as opposed to a signal because the client may want to wait for nodes to be allocated
// before sending work to those nodes.
// Returns the list of node names that were allocated to the job.
func (cm *ClusterManager) AssignNodesToJobs(ctx context.Context, input ClusterManagerAssignNodesToJobInput) (ClusterManagerAssignNodesToJobResult, error) {
	select {
	case <-ctx.Done():
		return ClusterManagerAssignNodesToJobResult{}, ctx.Err()
	case <-cm.started:
	}
	if cm.state.ClusterShutdown {
		// If you want the client to receive a error, either add an update validator and return the
		// error from there, or return an error.
		return ClusterManagerAssignNodesToJobResult{}, errors.New("cannot assign nodes to a job: Cluster is already shut down")
	}
	cm.nodeLock.Lock()
	defer cm.nodeLock.Unlock()
	// Idempotency guard.
	if _, ok := cm.state.JobsAssigned[input.JobName]; ok {
		return ClusterManagerAssignNodesToJobResult{
			NodesAssigned: cm.getAssignedNodes(&input.JobName),
		}, nil
	}
	unassignedNodes := cm.getUnassignedNodes()
	if len(unassignedNodes) < input.TotalNumNodes {
		return ClusterManagerAssignNodesToJobResult{}, errors.New("not enough nodes to assign to job")
	}
	nodesToAssign := unassignedNodes[:input.TotalNumNodes]

	// Assign the nodes to the job.
	activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Second * 10,
	})
	// This get would be dangerous without holding nodeLock because it yields control and allows interleaving
	// with DeleteJob and performHealthChecks, which both modify cm.state.nodes.
	err := workflow.ExecuteActivity(activityCtx, AssignNodesToJobsActivity, AssignNodesToJobInput{
		Nodes:   nodesToAssign,
		JobName: input.JobName,
	}).Get(activityCtx, nil)
	if err != nil {
		return ClusterManagerAssignNodesToJobResult{}, err
	}
	for _, node := range nodesToAssign {
		cm.state.Nodes[node] = input.JobName
	}
	cm.state.JobsAssigned[input.JobName] = struct{}{}

	return ClusterManagerAssignNodesToJobResult{
		NodesAssigned: cm.getAssignedNodes(&input.JobName),
	}, nil
}

// Even though it returns nothing, this is an update because the client may want to track it, for example
// to wait for nodes to be unassigned before reassigning them.
func (cm *ClusterManager) DeleteJob(ctx context.Context, input ClusterManagerDeleteJobInput) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-cm.started:
	}
	if cm.state.ClusterShutdown {
		// If you want the client to receive a error, either add an update validator and return the
		// error from there, or return an error.
		return errors.New("cannot delete a job: Cluster is already shut down")
	}
	cm.nodeLock.Lock()
	defer cm.nodeLock.Unlock()

	nodesToUnassign := make([]string, 0)
	for k := range cm.state.Nodes {
		if cm.state.Nodes[k] == input.JobName {
			nodesToUnassign = append(nodesToUnassign, k)
		}
	}
	activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Second * 10,
	})
	err := workflow.ExecuteActivity(activityCtx, UnassignNodesForJobActivity, UnassignNodesForJobInput{
		Nodes:   nodesToUnassign,
		JobName: input.JobName,
	}).Get(activityCtx, nil)
	if err != nil {
		return err
	}
	for _, node := range nodesToUnassign {
		cm.state.Nodes[node] = ""
	}
	return nil
}

func (cm *ClusterManager) getUnassignedNodes() []string {
	var unassignedNodes []string
	for k := range cm.state.Nodes {
		if cm.state.Nodes[k] == "" {
			unassignedNodes = append(unassignedNodes, k)
		}
	}
	return unassignedNodes

}

func (cm *ClusterManager) performHealthCheck(ctx context.Context) {
	cm.nodeLock.Lock()
	defer cm.nodeLock.Unlock()
	assignedNodes := cm.getAssignedNodes(nil)
	// Do some activity on the nodes
	var badNodes map[string]struct{}
	activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 10,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})
	err := workflow.ExecuteActivity(activityCtx, FindBadNodesActivity, assignedNodes).Get(activityCtx, &badNodes)
	if err != nil {
		cm.logger.Error("Health check failed", "error", err)
	}
	for node := range badNodes {
		cm.state.Nodes[node] = "BAD!"
	}
}

func (cm *ClusterManager) getAssignedNodes(jobName *string) map[string]struct{} {
	assignedNodes := make(map[string]struct{})
	if jobName != nil {
		for k := range cm.state.Nodes {
			if cm.state.Nodes[k] == *jobName {
				assignedNodes[k] = struct{}{}
			}
		}
	} else {
		for k := range cm.state.Nodes {
			if cm.state.Nodes[k] != "BAD!" {
				assignedNodes[k] = struct{}{}
			}
		}
	}
	return assignedNodes
}

func (cm *ClusterManager) shouldContinueAsNew(ctx context.Context) bool {
	if workflow.GetInfo(ctx).GetContinueAsNewSuggested() {
		return true
	}
	if cm.maxHistoryLength > 0 && workflow.GetInfo(ctx).GetCurrentHistoryLength() > cm.maxHistoryLength {
		return true
	}
	return false
}

func (cm *ClusterManager) run(ctx context.Context) (ClusterManagerResult, error) {
	if !cm.state.ClusterStarted {
		select {
		case <-ctx.Done():
			return ClusterManagerResult{}, ctx.Err()
		case <-cm.startCh:
		}
	}
	cm.state.ClusterStarted = true
	close(cm.started)
	timer := time.NewTimer(cm.sleepInterval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return ClusterManagerResult{}, ctx.Err()
		case <-cm.shutdownCh:
			cm.state.ClusterShutdown = true
		case <-timer.C:
			cm.performHealthCheck(ctx)
			timer.Reset(cm.sleepInterval)
		}
		if cm.state.ClusterShutdown {
			if err := workflow.WaitHandlersFinished(ctx); err != nil {
				return ClusterManagerResult{}, err
			}
			return ClusterManagerResult{NumCurrentlyAssignedNodes: len(cm.getAssignedNodes(nil)), NumBadNodes: len(cm.badNodes())}, nil
		}
		if cm.shouldContinueAsNew(ctx) {
			if err := workflow.WaitHandlersFinished(ctx); err != nil {
				return ClusterManagerResult{}, err
			}
			return ClusterManagerResult{}, workflow.NewContinueAsNewError(ctx, ClusterManagerWorkflow, ClusterManagerInput{State: &cm.state, TestContinueAsNew: cm.testContinueAsNew})
		}
	}
}

// ClusterManagerWorkflow keeps track of the assignments of a cluster of nodes.
// Via signals, the cluster can be started and shutdown.
// Via updates, clients can also assign jobs to nodes and delete jobs.
// These updates must run atomically.
//
//go:isolate
func ClusterManagerWorkflow(ctx context.Context, wfInput ClusterManagerInput) (ClusterManagerResult, error) {
	cm, err := newClusterManager(ctx, wfInput)
	if err != nil {
		return ClusterManagerResult{}, err
	}
	return cm.run(ctx)
}

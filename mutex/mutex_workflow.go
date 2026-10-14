package mutex

import (
	"context"
	"fmt"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

const (
	// AcquireLockSignalName signal channel name for lock acquisition
	AcquireLockSignalName = "acquire-lock-event"
	// RequestLockSignalName channel name for request lock
	RequestLockSignalName = "request-lock-event"

	ClientContextKey ContextKey = "Client"
)

type (
	ContextKey string

	UnlockFunc func() error

	Mutex struct {
		currentWorkflowID string
		lockNamespace     string
	}
)

// NewMutex initializes mutex
func NewMutex(currentWorkflowID string, lockNamespace string) *Mutex {
	return &Mutex{
		currentWorkflowID: currentWorkflowID,
		lockNamespace:     lockNamespace,
	}
}

// Lock - locks mutex
func (s *Mutex) Lock(ctx context.Context,
	resourceID string, unlockTimeout time.Duration) (UnlockFunc, error) {

	activityCtx := workflow.WithLocalActivityOptions(ctx, workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5,
		},
	})

	var releaseLockChannelName string
	var execution workflow.Execution
	err := workflow.ExecuteLocalActivity(activityCtx,
		SignalWithStartMutexWorkflowActivity, s.lockNamespace,
		resourceID, s.currentWorkflowID, unlockTimeout).Get(ctx, &execution)
	if err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case releaseLockChannelName = <-workflow.GetSignalChannel[string](ctx, AcquireLockSignalName):
	}

	unlockFunc := func() error {
		return workflow.SignalExternalWorkflow(ctx, execution.ID, execution.RunID,
			releaseLockChannelName, "releaseLock").Get(ctx, nil)
	}
	return unlockFunc, nil
}

// MutexWorkflow used for locking a resource
//
//temporal:workflow
func MutexWorkflow(
	ctx context.Context,
	namespace string,
	resourceID string,
	unlockTimeout time.Duration,
) error {
	currentWorkflowID := workflow.GetInfo(ctx).WorkflowExecution.ID

	logger := workflow.GetLogger(ctx)
	logger.Info("started", "currentWorkflowID", currentWorkflowID)
	requestLockCh := workflow.GetSignalChannel[string](ctx, RequestLockSignalName)
	for {
		var senderWorkflowID string
		select {
		case <-ctx.Done():
			return ctx.Err()
		case senderWorkflowID = <-requestLockCh:
		default:
			logger.Info("no more signals")
			return nil
		}
		releaseLockChannelName := generateUnlockChannelName(senderWorkflowID)
		logger.Info("generated release lock channel name", "releaseLockChannelName", releaseLockChannelName)
		// Send release lock channel name back to a senderWorkflowID, so that it can
		// release the lock using release lock channel name
		err := workflow.SignalExternalWorkflow(ctx, senderWorkflowID, "",
			AcquireLockSignalName, releaseLockChannelName).Get(ctx, nil)
		if err != nil {
			// .Get(ctx, nil) blocks until the signal is sent.
			// If the senderWorkflowID is closed (terminated/canceled/timeouted/completed/etc), this would return error.
			// In this case we release the lock immediately instead of failing the mutex workflow.
			// Mutex workflow failing would lead to all workflows that have sent requestLock will be waiting.
			logger.Info("SignalExternalWorkflow error", "Error", err)
			continue
		}
		logger.Info("signaled external workflow")
		release := workflow.GetSignalChannel[string](ctx, releaseLockChannelName)
		timer := time.NewTimer(unlockTimeout)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			logger.Info("unlockTimeout exceeded")
		case <-release:
			logger.Info("release signal received")
		}
		timer.Stop()
	}
}

// SignalWithStartMutexWorkflowActivity ...
func SignalWithStartMutexWorkflowActivity(
	ctx context.Context,
	namespace string,
	resourceID string,
	senderWorkflowID string,
	unlockTimeout time.Duration,
) (*workflow.Execution, error) {

	c := ctx.Value(ClientContextKey).(client.Client)
	workflowID := fmt.Sprintf(
		"%s:%s:%s",
		"mutex",
		namespace,
		resourceID,
	)
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "mutex",
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5,
		},
	}
	wr, err := c.SignalWithStartWorkflow(
		ctx, workflowID, RequestLockSignalName, senderWorkflowID,
		workflowOptions, "MutexWorkflow", namespace, resourceID, unlockTimeout)

	if err != nil {
		activity.GetLogger(ctx).Error("Unable to signal with start workflow", "Error", err)
		return nil, err
	}
	activity.GetLogger(ctx).Info("Signaled and started Workflow", "WorkflowID", wr.GetID(), "RunID", wr.GetRunID())

	return &workflow.Execution{
		ID:    wr.GetID(),
		RunID: wr.GetRunID(),
	}, nil
}

// generateUnlockChannelName generates release lock channel name
func generateUnlockChannelName(senderWorkflowID string) string {
	return fmt.Sprintf("unlock-event-%s", senderWorkflowID)
}

//temporal:workflow
func SampleWorkflowWithMutex(
	ctx context.Context,
	resourceID string,
) error {
	currentWorkflowID := workflow.GetInfo(ctx).WorkflowExecution.ID
	logger := workflow.GetLogger(ctx)
	logger.Info("started", "currentWorkflowID", currentWorkflowID, "resourceID", resourceID)

	mutex := NewMutex(currentWorkflowID, "TestUseCase")
	unlockFunc, err := mutex.Lock(ctx, resourceID, 10*time.Minute)
	if err != nil {
		return err
	}
	logger.Info("resource locked")

	// emulate long running process
	logger.Info("critical operation started")
	_ = sleep(ctx, 10*time.Second)
	logger.Info("critical operation finished")

	_ = unlockFunc()

	logger.Info("finished")
	return nil
}

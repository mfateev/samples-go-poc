package batch_sliding_window

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
	enumspb "go.temporal.io/api/enums/v1"
)

type (
	// SlidingWindowWorkflowInput contains SlidingWindowWorkflow arguments
	SlidingWindowWorkflowInput struct {
		PageSize          int
		SlidingWindowSize int
		Offset            int // inclusive
		MaximumOffset     int // exclusive
		Progress          int
		// The set of ids
		CurrentRecords map[int]bool // recordId -> ignored boolean
	}

	// SlidingWindow structure that implements the workflow logic
	SlidingWindow struct {
		input SlidingWindowWorkflowInput
		// currentRecords represents a set of records that are currently being processed by child workflows.
		// key is recordId. values are ignored.
		currentRecords map[int]bool
		// childrenStartedByThisRun is used to wait for children to start before calling continue as new.
		childrenStartedByThisRun []workflow.ChildWorkflowFuture
		// Offset into the next record to process.
		offset int
		// Count of completed records.
		progress         int
		reportCompletion <-chan int
	}

	// SlidingWindowState used as a "state" query result.
	SlidingWindowState struct {
		// currentRecords represents a set of record ids that are currently being processed by child workflows.
		CurrentRecords           []int
		ChildrenStartedByThisRun int
		Offset                   int
		Progress                 int
	}
)

// SlidingWindowWorkflow workflow processes a range of records using a requested number of child workflows.
// As soon as a child workflow completes a new one is started.
//
//temporal:workflow
func SlidingWindowWorkflow(ctx context.Context, input SlidingWindowWorkflowInput) (recordCount int, err error) {
	workflow.GetLogger(ctx).Info("SlidingWindowWorkflow",
		"input", input.SlidingWindowSize,
		"PageSize", input.PageSize,
		"Offset", input.Offset,
		"MaximumOffset", input.MaximumOffset,
		"Progress", input.Progress)

	impl := &SlidingWindow{
		input:          input,
		currentRecords: input.CurrentRecords,
		offset:         input.Offset,
		progress:       input.Progress,
	}
	if impl.currentRecords == nil {
		impl.currentRecords = make(map[int]bool)
	}
	err = workflow.SetQueryHandler(ctx, "state", func() (SlidingWindowState, error) {
		return impl.State()
	})
	if err != nil {
		return 0, err
	}
	return impl.Execute(ctx)
}

// State returns the current state of the batch.
// Used by the "state" workflow query.
func (s *SlidingWindow) State() (SlidingWindowState, error) {
	currentRecordIds := make([]int, len(s.currentRecords))
	i := 0
	for k := range s.currentRecords {
		currentRecordIds[i] = k
		i++
	}
	sort.Ints(currentRecordIds)
	return SlidingWindowState{
		CurrentRecords:           currentRecordIds,
		ChildrenStartedByThisRun: len(s.childrenStartedByThisRun),
		Offset:                   s.offset,
		Progress:                 s.progress,
	}, nil
}

func (s *SlidingWindow) Execute(ctx context.Context) (recordCount int, err error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
	})

	if s.input.PageSize <= 0 || s.input.SlidingWindowSize <= 0 {
		return 0, fmt.Errorf("page and sliding window sizes must be positive")
	}
	s.reportCompletion = workflow.GetSignalChannel[int](ctx, "ReportCompletion")

	var getOutput GetRecordsOutput
	if s.offset < s.input.MaximumOffset {
		getInput := GetRecordsInput{
			PageSize:  s.input.PageSize,
			Offset:    s.offset,
			MaxOffset: s.input.MaximumOffset,
		}
		var loader *RecordLoader
		err = workflow.ExecuteActivity(ctx, loader.GetRecords, getInput).Get(ctx, &getOutput)
		if err != nil {
			return 0, err
		}
	}
	workflowId := workflow.GetInfo(ctx).WorkflowExecution.ID
	// Process records
	for _, record := range getOutput.Records {
		// Blocks until the total number of children (including started by the previous runs)
		// gets below the SlidingWindowSize.
		for len(s.currentRecords) >= s.input.SlidingWindowSize {
			if err := s.waitForCompletion(ctx); err != nil {
				return 0, err
			}
		}

		options := workflow.ChildWorkflowOptions{
			// Use ABANDON as child workflows have to survive the parent calling continue-as-new
			ParentClosePolicy: enumspb.PARENT_CLOSE_POLICY_ABANDON,
			// Human readable child id.
			WorkflowID: fmt.Sprintf("%s/%d", workflowId, record.Id),
		}
		childCtx := workflow.WithChildWorkflowOptions(ctx, options)
		child := workflow.ExecuteChildWorkflow(childCtx, RecordProcessorWorkflow, record)

		s.childrenStartedByThisRun = append(s.childrenStartedByThisRun, child)
		s.currentRecords[record.Id] = true // value is ignored
		s.offset++
	}
	return s.continueAsNewOrComplete(ctx)
}

func (s *SlidingWindow) continueAsNewOrComplete(ctx context.Context) (int, error) {
	// Continues-as-new after starting PageSize children
	if s.offset < s.input.MaximumOffset {
		// Waits for all children to start. Without this wait, workflow completion through
		// continue-as-new might lead to a situation when they never start.
		for _, child := range s.childrenStartedByThisRun {
			err := child.GetChildWorkflowExecution().Get(ctx, nil)
			// Is not expected as children's automatically generated
			// IDs are not expected to collide.
			if err != nil {
				return 0, err
			}
		}
		// Must drain the signal channel without blocking before calling continue-as-new.
		// Failure to do so can lead to signal loss.
		if err := s.drainCompletionSignals(ctx); err != nil {
			return 0, err
		}

		// Returns ContinueAsNewError with new workflow input
		newInput := SlidingWindowWorkflowInput{
			PageSize:          s.input.PageSize,
			SlidingWindowSize: s.input.SlidingWindowSize,
			Offset:            s.offset,
			MaximumOffset:     s.input.MaximumOffset,
			Progress:          s.progress,
			CurrentRecords:    s.currentRecords,
		}
		return 0, workflow.NewContinueAsNewError(ctx, SlidingWindowWorkflow, newInput)
	}
	// The final run waits for every record, including children from earlier runs.
	for len(s.currentRecords) != 0 {
		if err := s.waitForCompletion(ctx); err != nil {
			return 0, err
		}
	}
	return s.progress, nil
}

func (s *SlidingWindow) waitForCompletion(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case recordID, ok := <-s.reportCompletion:
		if !ok {
			return ctx.Err()
		}
		s.recordCompletion(recordID)
		return nil
	}
}

func (s *SlidingWindow) drainCompletionSignals(ctx context.Context) error {
	for {
		select {
		case recordID, ok := <-s.reportCompletion:
			if !ok {
				return ctx.Err()
			}
			s.recordCompletion(recordID)
		default:
			return ctx.Err()
		}
	}
}

func (s *SlidingWindow) recordCompletion(recordID int) {
	// A retried notification must not increment progress twice.
	if s.currentRecords[recordID] {
		delete(s.currentRecords, recordID)
		s.progress++
	}
}

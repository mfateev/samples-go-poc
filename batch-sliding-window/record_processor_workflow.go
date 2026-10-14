package batch_sliding_window

import (
	"context"
	"math/rand"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
)

// RecordProcessorWorkflow workflow that implements processing of a single record.
//
//temporal:workflow
func RecordProcessorWorkflow(ctx context.Context, r SingleRecord) error {
	err := ProcessRecord(ctx, r)
	// Notify parent about completion via signal
	parent := workflow.GetInfo(ctx).ParentWorkflowExecution
	// This workflow is always expected to have a parent.
	// But for unit testing it might be useful to skip the notification if there is none.
	if parent != nil {
		// Doesn't specify runId as parent calls continue-as-new.
		signaled := workflow.SignalExternalWorkflow(ctx, parent.ID, "", "ReportCompletion", r.Id)
		// Ensure that signal is delivered.
		// Completing workflow before this Future is ready might lead to the signal loss.
		signalErr := signaled.Get(ctx, nil)
		if signalErr != nil {
			return signalErr
		}
	}
	return err
}

// ProcessRecord simulates application specific record processing.
func ProcessRecord(ctx context.Context, r SingleRecord) error {
	// Simulate some processing

	// The isolate's native random stream is deterministic during replay.
	err := sleep(ctx, time.Duration(rand.Intn(10))*time.Second)
	if err != nil {
		return err
	}
	workflow.GetLogger(ctx).Info("Processed ", r)
	return nil
}

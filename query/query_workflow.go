package query

import (
	"time"

	"context"

	"github.com/mfateev/sdk-go-poc/workflow"
)

// Workflow is to demo how to setup query handler
//
//temporal:workflow
func QueryWorkflow(ctx context.Context) error {
	queryResult := "started"
	logger := workflow.GetLogger(ctx)
	logger.Info("QueryWorkflow started")
	// setup query handler for query type "state"
	err := workflow.SetQueryHandler(ctx, "state", func() (string, error) {
		return queryResult, nil
	})
	if err != nil {
		logger.Info("SetQueryHandler failed: " + err.Error())
		return err
	}

	queryResult = "waiting on timer"
	// to simulate workflow been blocked on something, in reality, workflow could wait on anything like activity, signal or timer
	_ = sleep(ctx, time.Minute*2)
	logger.Info("Timer fired")

	queryResult = "done"
	logger.Info("QueryWorkflow completed")
	return nil
}

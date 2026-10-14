package memo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/log"
)

/*
 * This sample shows how to use memo.
 */

// ClientContextKey is the key for lookup
type ClientContextKey struct{}

var (
	// ErrClientNotFound when client is not found on context.
	ErrClientNotFound = errors.New("failed to retrieve client from context")
	// ClientCtxKey for retrieving client from context.
	ClientCtxKey ClientContextKey
)

// MemoWorkflow workflow definition
//
//temporal:workflow
func MemoWorkflow(ctx context.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Memo workflow started")

	// Get memo that were provided when workflow was started.
	info := workflow.GetInfo(ctx)
	val := info.Memo.Fields["description"]
	var currentDescription string
	err := converter.GetDefaultDataConverter().FromPayload(val, &currentDescription)
	if err != nil {
		logger.Error("Get memo failed.", "Error", err)
		return err
	}
	logger.Info("Current memo value.", "description", currentDescription)

	// Update memo.
	memo := map[string]any{
		"description": "Test upsert memo workflow",
	}
	// This won't persist memo on server because commands are not sent to server,
	// but local cache will be updated.
	err = workflow.UpsertMemo(ctx, memo)
	if err != nil {
		return err
	}

	// Print current memo with modifications above.
	info = workflow.GetInfo(ctx)
	err = printMemo(info.Memo, logger)
	if err != nil {
		return err
	}

	// Now send commands to the server and let visibility storage update the index.
	_ = sleep(ctx, 1*time.Second)

	// After visibility storage index is updated we can query it.
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	var visibleMemo map[string]any
	err = workflow.ExecuteActivity(ctx, DescribeWorkflow, info.WorkflowExecution.ID).Get(ctx, &visibleMemo)
	if err != nil {
		logger.Error("Failed to describe workflow execution.", "Error", err)
		return err
	}

	// Print current memo from visibility storage.
	logger.Info("Current memo from visibility storage", "memo", visibleMemo)

	logger.Info("Workflow completed.")
	return nil
}

func printMemo(memo *common.Memo, logger log.Logger) error {
	if memo == nil || len(memo.GetFields()) == 0 {
		logger.Info("Current memo is empty.")
		return nil
	}

	var builder strings.Builder
	for k, v := range memo.GetFields() {
		var currentVal any
		err := converter.GetDefaultDataConverter().FromPayload(v, &currentVal)
		if err != nil {
			logger.Error(fmt.Sprintf("Get memo for key %s failed.", k), "Error", err)
			return err
		}
		builder.WriteString(fmt.Sprintf("%s=%v\n", k, currentVal))
	}
	logger.Info(fmt.Sprintf("Current memo values:\n%s", builder.String()))
	return nil
}

func DescribeWorkflow(ctx context.Context, wfID string) (map[string]any, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Describe workflow.", "WorkflowId", wfID)

	c, err := getClientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := c.DescribeWorkflowExecution(ctx, wfID, "")
	if err != nil {
		return nil, err
	}

	memo := make(map[string]any)
	for key, payload := range resp.GetWorkflowExecutionInfo().GetMemo().GetFields() {
		var value any
		if err := converter.GetDefaultDataConverter().FromPayload(payload, &value); err != nil {
			return nil, err
		}
		memo[key] = value
	}
	return memo, nil
}

func getClientFromContext(ctx context.Context) (client.Client, error) {
	logger := activity.GetLogger(ctx)
	c, ok := ctx.Value(ClientCtxKey).(client.Client)
	if c == nil || !ok {
		logger.Error("Could not retrieve client from context.")
		return nil, ErrClientNotFound
	}

	return c, nil
}

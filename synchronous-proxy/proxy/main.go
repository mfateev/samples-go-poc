package proxy

import (
	"fmt"

	"context"

	"github.com/mfateev/sdk-go-poc/workflow"
)

const proxyRequestSignalName = "proxy-request-signal"
const proxyResponseSignalName = "proxy-response-signal"

type proxySignalRequest struct {
	Key               string
	Value             string
	CallingWorkflowId string
}

type proxySignalResponse struct {
	Key   string
	Value string
	Error string
}

func SendErrorResponse(ctx context.Context, id string, err error) error {
	logger := workflow.GetLogger(ctx)

	logger.Info("Sending error response", id)

	return workflow.SignalExternalWorkflow(
		ctx,
		id,
		"",
		proxyResponseSignalName,
		proxySignalResponse{Error: err.Error()},
	).Get(ctx, nil)
}

func SendResponse(ctx context.Context, id string, key string, value string) error {
	logger := workflow.GetLogger(ctx)

	logger.Info("Sending response", id)

	return workflow.SignalExternalWorkflow(
		ctx,
		id,
		"",
		proxyResponseSignalName,
		proxySignalResponse{Key: key, Value: value},
	).Get(ctx, nil)
}

func SendRequest(ctx context.Context, targetWorkflowID string, key string, value string) error {
	logger := workflow.GetLogger(ctx)

	workflowID := workflow.GetInfo(ctx).WorkflowExecution.ID

	logger.Info("Sending request", targetWorkflowID, workflowID)

	return workflow.SignalExternalWorkflow(
		ctx,
		targetWorkflowID,
		"",
		proxyRequestSignalName,
		proxySignalRequest{
			CallingWorkflowId: workflowID,
			Key:               key,
			Value:             value,
		},
	).Get(ctx, nil)
}

func ReceiveResponse(ctx context.Context) (string, string, error) {
	logger := workflow.GetLogger(ctx)

	var res proxySignalResponse

	ch := workflow.GetSignalChannel[proxySignalResponse](ctx, proxyResponseSignalName)

	logger.Info("Waiting for response")

	select {
	case <-ctx.Done():
		return "", "", ctx.Err()
	case value, ok := <-ch:
		if !ok {
			return "", "", ctx.Err()
		}
		res = value
	}

	logger.Info("Received response")

	if res.Error != "" {
		return "", "", fmt.Errorf("%s", res.Error)
	}

	return res.Key, res.Value, nil
}

func ReceiveRequest(ctx context.Context) (string, string, string) {
	logger := workflow.GetLogger(ctx)

	var req proxySignalRequest

	ch := workflow.GetSignalChannel[proxySignalRequest](ctx, proxyRequestSignalName)

	logger.Info("Waiting for request")

	select {
	case <-ctx.Done():
		return "", "", ""
	case value, ok := <-ch:
		if !ok {
			return "", "", ""
		}
		req = value
	}

	logger.Info("Received request")

	return req.CallingWorkflowId, req.Key, req.Value
}

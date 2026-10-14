package periodic_sequence

import (
	"time"

	"context"

	"github.com/mfateev/sdk-go-poc/workflow"
)

const (
	TaskQueueName = "pollingPeriodicSequenceSampleQueue"
)

//temporal:workflow
func PeriodicSequencePolling(ctx context.Context, pollingInterval time.Duration) (string, error) {
	cwo := workflow.ChildWorkflowOptions{}
	ctx = workflow.WithChildWorkflowOptions(ctx, cwo)
	params := ChildWorkflowParams{
		PollingInterval: pollingInterval,
	}
	res := workflow.ExecuteChildWorkflow(ctx, PollingChildWorkflow, params)
	var result string
	err := res.Get(ctx, &result)
	return result, err

}

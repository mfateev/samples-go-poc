package logger_interceptor

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mfateev/sdk-go-poc/interceptor"
	"github.com/mfateev/sdk-go-poc/workflow"
	"go.temporal.io/sdk/log"
)

//go:isolate
func NewWorkflowInterceptors(config []byte) ([]interceptor.WorkerInterceptor, error) {
	var tags []any
	if len(config) > 0 {
		if err := json.Unmarshal(config, &tags); err != nil {
			return nil, err
		}
	}
	return []interceptor.WorkerInterceptor{&nativeWorker{tags: tags}}, nil
}

type nativeWorker struct {
	tags []any
	interceptor.WorkerInterceptorBase
}

func (w *nativeWorker) InterceptWorkflow(ctx context.Context, next interceptor.WorkflowInboundInterceptor) interceptor.WorkflowInboundInterceptor {
	i := &nativeInbound{tags: w.tags}
	i.Next = next
	return i
}

type nativeInbound struct {
	tags []any
	interceptor.WorkflowInboundInterceptorBase
}

func (i *nativeInbound) Init(next interceptor.WorkflowOutboundInterceptor) error {
	o := &nativeOutbound{tags: i.tags}
	o.Next = next
	return i.Next.Init(o)
}

type nativeOutbound struct {
	tags []any
	interceptor.WorkflowOutboundInterceptorBase
}

func (o *nativeOutbound) GetLogger(ctx context.Context) log.Logger {
	if len(o.tags) > 0 {
		return log.With(o.Next.GetLogger(ctx), o.tags...)
	}
	return log.With(o.Next.GetLogger(ctx), "WorkflowStartTime", workflow.GetInfo(ctx).WorkflowStartTime.Format(time.RFC3339))
}

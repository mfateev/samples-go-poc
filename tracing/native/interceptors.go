// Package native constructs private tracing interceptors inside workflows.
package native

import (
	dd "github.com/mfateev/sdk-go-poc/contrib/datadog/tracing"
	otelv1 "github.com/mfateev/sdk-go-poc/contrib/opentelemetry"
	otelv2 "github.com/mfateev/sdk-go-poc/contrib/opentelemetry-v2"
	"github.com/mfateev/sdk-go-poc/interceptor"
	"github.com/mfateev/sdk-go-poc/interceptor/tracing"
)

const SinkOp uint32 = 1000

//go:isolate
func NewInterceptors(config []byte) ([]interceptor.WorkerInterceptor, error) {
	switch string(config) {
	case "datadog":
		return []interceptor.WorkerInterceptor{dd.NewTracingInterceptor(dd.TracerOptions{SinkOp: SinkOp})}, nil
	case "v2", "v2-spans":
		return []interceptor.WorkerInterceptor{otelv2.NewTracingInterceptor(otelv2.TracerOptions{SinkOp: SinkOp, TracerOptions: tracing.TracerOptions{AddTemporalSpans: string(config) == "v2-spans"}})}, nil
	default:
		i, err := otelv1.NewTracingInterceptor(otelv1.TracerOptions{SinkOp: SinkOp})
		return []interceptor.WorkerInterceptor{i}, err
	}
}

# OpenTelemetry v2

These samples use the
[OpenTelemetry v2 plugin](https://pkg.go.dev/go.temporal.io/sdk/contrib/opentelemetry-v2@v0.1.0)
for replay-safe tracing and metrics.

## Jaeger

These samples export traces through OpenTelemetry to
[Jaeger, an OpenTelemetry-compatible tracing backend](https://www.jaegertracing.io/docs/latest/#relationship-with-opentelemetry).
Jaeger visualizes each trace as a timeline showing how an operation moved
through your code.

## Samples

- [Automatic instrumentation](automatic-instrumentation)
- [Workflow-to-Activity propagation](workflow-activity-propagation)
- [Client-to-Update propagation](client-update-propagation)

Shared host setup lives in [`setup/setup.go`](setup/setup.go). Native workflow
interceptor factories and copied span export use the POC tracing integration.

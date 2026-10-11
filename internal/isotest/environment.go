// Package isotest runs the upstream SDK test environment with compiled native
// workflow definitions. The test driver must be built with the isolate compiler.
package isotest

import (
	"context"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/stretchr/testify/mock"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	bindings "go.temporal.io/sdk/internalbindings"

	isolateworker "github.com/mfateev/sdk-go-poc/worker"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

type WorkflowTestSuite struct{ testsuite.WorkflowTestSuite }
type TestUpdateCallback = testsuite.TestUpdateCallback
type TestActivityEnvironment = testsuite.TestActivityEnvironment
type DevServerOptions = testsuite.DevServerOptions

var StartDevServer = testsuite.StartDevServer

var ErrMockStartChildWorkflowFailed = testsuite.ErrMockStartChildWorkflowFailed
var workflows []any

func RegisterWorkflows(entries ...any) { workflows = entries }

type TestWorkflowEnvironment struct {
	*testsuite.TestWorkflowEnvironment
	registrar worker.Worker
	names     map[uintptr]string
	functions map[string]any
}
type backend struct {
	worker.Worker
	env *testsuite.TestWorkflowEnvironment
}

var retainedDefinitions []bindings.WorkflowDefinition

// CloseEnvironments releases definitions retained for post-completion queries.
func CloseEnvironments() {
	for _, definition := range retainedDefinitions {
		definition.Close()
	}
	retainedDefinitions = nil
}

func (b backend) RegisterWorkflowWithOptions(fn any, opts workflow.RegisterOptions) {
	if f, ok := fn.(bindings.WorkflowDefinitionFactory); ok {
		fn = patientFactory{f}
	}
	b.env.RegisterWorkflowWithOptions(fn, opts)
}
func (b backend) RegisterActivityWithOptions(fn any, opts activity.RegisterOptions) {
	b.env.RegisterActivityWithOptions(fn, opts)
}
func (s *WorkflowTestSuite) NewTestWorkflowEnvironment() *TestWorkflowEnvironment {
	env := s.WorkflowTestSuite.NewTestWorkflowEnvironment()
	result := &TestWorkflowEnvironment{TestWorkflowEnvironment: env, registrar: isolateworker.Wrap(backend{env: env}), names: make(map[uintptr]string), functions: make(map[string]any)}
	for _, fn := range workflows {
		result.RegisterWorkflow(fn)
	}
	return result
}
func functionName(fn any) string {
	name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	name = strings.TrimSuffix(name, "-fm")
	return name[strings.LastIndex(name, ".")+1:]
}
func (e *TestWorkflowEnvironment) RegisterWorkflow(fn any) {
	e.RegisterWorkflowWithOptions(fn, workflow.RegisterOptions{})
}
func (e *TestWorkflowEnvironment) RegisterWorkflowWithOptions(fn any, opts workflow.RegisterOptions) {
	if opts.Name == "" {
		opts.Name = functionName(fn)
	}
	ptr := reflect.ValueOf(fn).Pointer()
	if previous, ok := e.names[ptr]; ok && previous == opts.Name {
		return
	}
	opts.DisableAlreadyRegisteredCheck = true
	e.registrar.RegisterWorkflowWithOptions(fn, opts)
	e.names[ptr] = opts.Name
	e.functions[opts.Name] = fn
}
func (e *TestWorkflowEnvironment) RegisterActivity(fn any) { e.registrar.RegisterActivity(fn) }
func (e *TestWorkflowEnvironment) RegisterActivityWithOptions(fn any, opts activity.RegisterOptions) {
	e.registrar.RegisterActivityWithOptions(fn, opts)
}
func (e *TestWorkflowEnvironment) ExecuteWorkflow(fn any, args ...any) {
	if _, ok := fn.(string); !ok {
		ptr := reflect.ValueOf(fn).Pointer()
		if _, ok := e.names[ptr]; !ok {
			e.RegisterWorkflow(fn)
		}
		fn = e.names[ptr]
	}
	e.TestWorkflowEnvironment.ExecuteWorkflow(fn, args...)
}

type patientFactory struct {
	bindings.WorkflowDefinitionFactory
}

func (f patientFactory) NewWorkflowDefinition() bindings.WorkflowDefinition {
	d := f.WorkflowDefinitionFactory.NewWorkflowDefinition()
	retainedDefinitions = append(retainedDefinitions, d)
	return patientDefinition{d}
}

type patientDefinition struct{ bindings.WorkflowDefinition }

// The SDK suite closes definitions immediately on completion. Keep native
// query state until the test finishes, matching a worker's retained cache.
func (d patientDefinition) Close() {}

// SDK unit environments omit OriginalRunID. Native sessions use that history
// identity, so supply the unit environment's stable run ID for these tests.
type testEnvironment struct{ bindings.WorkflowEnvironment }

func (e testEnvironment) WorkflowInfo() *workflow.Info {
	info := *e.WorkflowEnvironment.WorkflowInfo()
	if info.OriginalRunID == "" {
		info.OriginalRunID = info.WorkflowExecution.RunID
	}
	return &info
}

func (d patientDefinition) Execute(env bindings.WorkflowEnvironment, header *commonpb.Header, input *commonpb.Payloads) {
	d.WorkflowDefinition.Execute(testEnvironment{env}, header, input)
}

func (d patientDefinition) OnWorkflowTaskStarted(deadline time.Duration) {
	if deadline == 0 {
		deadline = 15 * time.Second
	}
	d.WorkflowDefinition.OnWorkflowTaskStarted(deadline)
}

// Mocked child workflows execute on the host. Register a signature-compatible
// SDK stub so its normal mock dispatch can replace the isolate definition.
func (e *TestWorkflowEnvironment) OnWorkflow(fn any, args ...any) *testsuite.MockCallWrapper {
	var original any
	name, ok := fn.(string)
	if ok {
		original = e.functions[name]
	} else {
		original = fn
		name = e.names[reflect.ValueOf(fn).Pointer()]
		if name == "" {
			name = functionName(fn)
		}
	}
	if original != nil {
		typ := reflect.TypeOf(original)
		inputs := make([]reflect.Type, typ.NumIn())
		for i := range inputs {
			inputs[i] = typ.In(i)
		}
		inputs[0] = reflect.TypeFor[workflow.Context]()
		outputs := make([]reflect.Type, typ.NumOut())
		for i := range outputs {
			outputs[i] = typ.Out(i)
		}
		stub := reflect.MakeFunc(reflect.FuncOf(inputs, outputs, false), func([]reflect.Value) []reflect.Value { panic("unmocked host workflow stub: " + name) }).Interface()
		e.TestWorkflowEnvironment.RegisterWorkflowWithOptions(stub, workflow.RegisterOptions{Name: name, DisableAlreadyRegisteredCheck: true})
	}
	return e.TestWorkflowEnvironment.OnWorkflow(name, args...)
}
func (e *TestWorkflowEnvironment) OnActivity(fn any, args ...any) *testsuite.MockCallWrapper {
	typ := reflect.TypeOf(fn)
	if typ.Kind() == reflect.Func && typ.NumIn() > 0 && typ.In(0) == reflect.TypeFor[context.Context]() && len(args) == typ.NumIn()-1 {
		args = append([]any{mock.Anything}, args...)
	}
	return e.TestWorkflowEnvironment.OnActivity(fn, args...)
}
func (e *TestWorkflowEnvironment) SetIsolateInterceptors(options isolateworker.InterceptorOptions) error {
	return isolateworker.SetIsolateInterceptors(e.registrar, options)
}

// DecodeResult extracts factory update results, which cross the isolate boundary
// as raw payloads instead of sharing the handler's concrete Go object.
func DecodeResult[T any](value any) T {
	if v, ok := value.(T); ok {
		return v
	}
	var result T
	raw, ok := value.(converter.RawValue)
	if !ok {
		panic("update result is not a raw payload")
	}
	if err := converter.GetDefaultDataConverter().FromPayload(raw.Payload(), &result); err != nil {
		panic(err)
	}
	return result
}

func (e *TestWorkflowEnvironment) SetIsolateDataConverter(options isolateworker.DataConverterOptions) error {
	return isolateworker.SetIsolateDataConverter(e.registrar, options)
}

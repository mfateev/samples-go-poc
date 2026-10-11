package dsl

import (
	"sync"
	"time"

	"context"

	"github.com/mfateev/sdk-go-poc/workflow"
)

type (
	// Workflow is the type used to express the workflow definition. Variables are a map of valuables. Variables can be
	// used as input to Activity.
	Workflow struct {
		Variables map[string]string
		Root      Statement
	}

	// Statement is the building block of dsl workflow. A Statement can be a simple ActivityInvocation or it
	// could be a Sequence or Parallel.
	Statement struct {
		Activity *ActivityInvocation
		Sequence *Sequence
		Parallel *Parallel
	}

	// Sequence consist of a collection of Statements that runs in sequential.
	Sequence struct {
		Elements []*Statement
	}

	// Parallel can be a collection of Statements that runs in parallel.
	Parallel struct {
		Branches []*Statement
	}

	// ActivityInvocation is used to express invoking an Activity. The Arguments defined expected arguments as input to
	// the Activity, the result specify the name of variable that it will store the result as which can then be used as
	// arguments to subsequent ActivityInvocation.
	ActivityInvocation struct {
		Name      string
		Arguments []string
		Result    string
	}

	bindings struct {
		mu     sync.Mutex
		values map[string]string
	}
	executable interface {
		execute(ctx context.Context, bindings *bindings) error
	}
)

// SimpleDSLWorkflow workflow definition
//
//go:isolate
func SimpleDSLWorkflow(ctx context.Context, dslWorkflow Workflow) ([]byte, error) {
	bindings := &bindings{values: make(map[string]string)}
	for k, v := range dslWorkflow.Variables {
		bindings.values[k] = v
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	err := dslWorkflow.Root.execute(ctx, bindings)
	if err != nil {
		logger.Error("DSL Workflow failed.", "Error", err)
		return nil, err
	}

	logger.Info("DSL Workflow completed.")
	return nil, err
}

func (b *Statement) execute(ctx context.Context, bindings *bindings) error {
	if b.Parallel != nil {
		err := b.Parallel.execute(ctx, bindings)
		if err != nil {
			return err
		}
	}
	if b.Sequence != nil {
		err := b.Sequence.execute(ctx, bindings)
		if err != nil {
			return err
		}
	}
	if b.Activity != nil {
		err := b.Activity.execute(ctx, bindings)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a ActivityInvocation) execute(ctx context.Context, bindings *bindings) error {
	inputParam := makeInput(a.Arguments, bindings)
	var result string
	err := workflow.ExecuteActivity(ctx, a.Name, inputParam).Get(ctx, &result)
	if err != nil {
		return err
	}
	if a.Result != "" {
		bindings.mu.Lock()
		bindings.values[a.Result] = result
		bindings.mu.Unlock()
	}
	return nil
}

func (s Sequence) execute(ctx context.Context, bindings *bindings) error {
	for _, a := range s.Elements {
		err := a.execute(ctx, bindings)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p Parallel) execute(ctx context.Context, bindings *bindings) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	completed := make(chan error, len(p.Branches))
	for _, branch := range p.Branches {
		go func() { completed <- branch.execute(ctx, bindings) }()
	}
	for range p.Branches {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-completed:
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func makeInput(argNames []string, argsMap *bindings) []string {
	argsMap.mu.Lock()
	defer argsMap.mu.Unlock()
	var args []string
	for _, arg := range argNames {
		args = append(args, argsMap.values[arg])
	}
	return args
}

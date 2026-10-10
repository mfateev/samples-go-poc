// Exercise the real marked workflow with reordered host completions.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"isolate"
	"runtime"
	"slices"
	"time"

	"github.com/mfateev/samples-go-poc/goroutines"
	"github.com/mfateev/samples-go-poc/internal/testbridge"
	"github.com/mfateev/sdk-go-poc/workflow"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/proto"
)

func main() {
	handle, ok := isolate.LookupFunction(goroutines.GreetAll)
	if !ok {
		panic("missing marked workflow")
	}
	for _, procs := range []int{1, 2, 8} {
		runtime.GOMAXPROCS(procs)
		run(handle, nil, false)
		for repetition := 0; repetition < 5; repetition++ {
			run(handle, []string{"Ada", "Grace", "Linus"}, false)
			run(handle, []string{"Ada", "Grace", "Linus"}, true)
		}
	}
	fmt.Println("native goroutine fan-out, empty input, reordered completions, and errors passed")
}

func run(handle isolate.Handle, names []string, failed bool) {
	clock := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	i, err := isolate.New(isolate.Config{
		Deterministic: true,
		Program:       handle.Program(func() { _ = workflow.RunFunction(handle) }),
		InitialTime:   &clock,
		TimerOp:       workflow.OpSleep,
	})
	check(err)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		check(i.Kill(ctx))
	}()
	check(i.Start())
	deadline := time.After(5 * time.Second)
	awaits := make(map[uint64]*isolate.Command)
	next := func(op uint32) *isolate.Command {
		for {
			select {
			case command := <-i.Commands():
				if command != nil && command.Op == workflow.OpWorkflowCancel {
					continue
				}
				if command != nil && command.Op == workflow.OpCancellableCall {
					var request workflow.CallRequest
					if err := json.Unmarshal(command.Payload, &request); err != nil {
						panic(err)
					}
					command.Op, command.Payload = request.Op, request.Payload
				}
				if command != nil && command.Op == workflow.OpAwaitActivity {
					var id uint64
					check(json.Unmarshal(command.Payload, &id))
					awaits[id] = command
					if op != workflow.OpAwaitActivity {
						continue
					}
				}
				if command == nil || command.Op != op {
					panic(fmt.Sprintf("expected %d, got %+v", op, command))
				}
				return command
			case <-deadline:
				panic("workflow did not progress")
			}

		}
	}
	input, err := converter.GetDefaultDataConverter().ToPayloads(names)
	check(err)
	inputBytes, err := proto.Marshal(input)
	check(err)
	testbridge.ReplyStart(next(workflow.OpStartPayloads), workflow.PayloadStart{Name: handle.Name(), Payloads: inputBytes})
	var ids []uint64
	for _, name := range names {
		command := next(workflow.OpScheduleActivity)
		var request workflow.ActivityPayloadRequest
		check(json.Unmarshal(command.Payload, &request))
		var args commonpb.Payloads
		check(proto.Unmarshal(request.Payloads, &args))
		var input string
		check(converter.GetDefaultDataConverter().FromPayloads(&args, &input))
		if len(args.Payloads) != 1 {
			panic("wrong greeting argument count")
		}
		if request.Name != "Greet" || input != name || request.Options == nil || request.Options.StartToCloseTimeout != 10*time.Second {
			panic(fmt.Sprintf("unexpected activity %+v", request))
		}
		ids = append(ids, request.ID)
		command.Reply(nil, nil)
	}
	for len(awaits) < len(names) {
		next(workflow.OpAwaitActivity)
	}
	if len(ids) != 0 {
		check(i.Suspend())
		for index := len(ids) - 1; index >= 0; index-- {
			var cause error
			if failed && index == 1 {
				cause = errors.New("greeting failed")
			}
			result, err := converter.GetDefaultDataConverter().ToPayloads("Hello " + names[index] + "!")
			check(err)
			encoded, err := proto.Marshal(result)
			check(err)
			outcome := workflow.ActivityOutcome{Payloads: encoded}
			if cause != nil {
				outcome.Error = cause.Error()
			}
			awaits[ids[index]].Reply(marshal(outcome), nil)
		}
		check(i.Resume())
	}
	command := next(workflow.OpCompletePayloads)
	var completion workflow.PayloadCompletion
	check(json.Unmarshal(command.Payload, &completion))
	if failed {
		if completion.Error != "greeting failed" {
			panic(fmt.Sprintf("error = %q", completion.Error))
		}
	} else {
		if completion.Error != "" {
			panic(completion.Error)
		}
		var payloads commonpb.Payloads
		check(proto.Unmarshal(completion.Payloads, &payloads))
		var greetings []string
		check(converter.GetDefaultDataConverter().FromPayloads(&payloads, &greetings))
		want := make([]string, len(names))
		for index, name := range names {
			want[index] = "Hello " + name + "!"
		}
		if !slices.Equal(greetings, want) {
			panic(fmt.Sprintf("greetings = %v, want %v", greetings, want))
		}
	}
	command.Reply(nil, nil)
	check(i.Wait())
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
func marshal(value any) []byte { data, err := json.Marshal(value); check(err); return data }

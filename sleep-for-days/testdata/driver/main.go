// Drive the sample's real isolate calls without the Temporal task adapter.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"isolate"
	"os"
	"reflect"
	"time"

	sleepfordays "github.com/mfateev/samples-go-poc/sleep-for-days"
	"github.com/mfateev/sdk-go-poc/workflow"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/proto"
)

func main() {
	handle, ok := isolate.LookupFunction(sleepfordays.SleepForDays)
	if !ok || handle.Signature().NumIn() != 1 || handle.Signature().In(0) != reflect.TypeFor[context.Context]() {
		panic("expected a marked context-only workflow")
	}
	clock := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	i, err := isolate.New(isolate.Config{
		Deterministic: true,
		Program:       handle.Program(func() { _ = workflow.RunFunction(handle) }),
		InitialTime:   &clock,
		TimerOp:       workflow.OpSleep,
	})
	check(err)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		check(i.Kill(ctx))
	}()
	check(i.Start())
	deadline := time.After(10 * time.Second)
	next := func() *isolate.Command {
		for {
			select {
			case command, ok := <-i.Commands():
				if ok && command.Op == workflow.OpWorkflowCancel {
					continue
				}
				if ok && command.Op == workflow.OpCancellableCall {
					var request workflow.CallRequest
					check(json.Unmarshal(command.Payload, &request))
					command.Op, command.Payload = request.Op, request.Payload
				}
				if ok && command.Op == workflow.OpAwaitActivity {
					if os.Args[1] == "failed-email" {
						command.Reply(marshal(workflow.ActivityOutcome{Error: "email failed"}), nil)
					}
					continue
				}
				if !ok {
					panic("isolate stopped before completion")
				}
				return command
			case <-deadline:
				panic("workflow did not make progress")
			}

		}
	}
	start := next()
	if start.Op != workflow.OpStartPayloads {
		panic("expected workflow start")
	}
	start.Reply(marshal(workflow.PayloadStart{Name: handle.Name()}), nil)
	var signal *isolate.Command
	for round := 0; round < 3; round++ {
		var timer *isolate.Command
		email := false
		for !email || timer == nil || signal == nil {
			command := next()
			switch command.Op {
			case workflow.OpScheduleActivity:
				var request workflow.ActivityPayloadRequest
				check(json.Unmarshal(command.Payload, &request))
				var args commonpb.Payloads
				check(proto.Unmarshal(request.Payloads, &args))
				var message string
				check(converter.GetDefaultDataConverter().FromPayloads(&args, &message))
				if len(args.Payloads) != 1 {
					panic("wrong email argument count")
				}
				if email || request.Name != "SendEmail" || message != "Sleeping for 30 days" || request.Options == nil || request.Options.StartToCloseTimeout != 10*time.Second {
					panic("unexpected email request")
				}
				email = true
				command.Reply(nil, nil)
				// The result wait stays pending unless the failed-email scenario replies.
			case workflow.OpSleep:
				var duration time.Duration
				check(json.Unmarshal(command.Payload, &duration))
				if timer != nil || duration != 30*24*time.Hour {
					panic(fmt.Sprintf("unexpected timer duration: %v", duration))
				}
				timer = command
			case workflow.OpSignal:
				if signal != nil || string(command.Payload) != "complete" {
					panic("unexpected signal subscription")
				}
				signal = command
			default:
				panic(fmt.Sprintf("unexpected operation: %d", command.Op))
			}
		}
		if round < 2 {
			clock = clock.Add(30 * 24 * time.Hour)
			check(i.AdvanceTime(clock))
			timer.Reply(nil, nil)
		} else {
			// Complete before the third timer fires, leaving it pending.
			signal.Reply(marshal(workflow.Signal{Name: "complete"}), nil)
		}
	}
	completion := next()
	// The buffered signal proxy may request another signal before the workflow
	// consumes the first one. It must not schedule another email or timer.
	for completion.Op == workflow.OpSignal && string(completion.Payload) == "complete" {
		completion = next()
	}
	if completion.Op != workflow.OpCompletePayloads {
		panic(fmt.Sprintf("workflow scheduled operation %d after the completion signal", completion.Op))
	}
	var result workflow.PayloadCompletion
	check(json.Unmarshal(completion.Payload, &result))
	if result.Error != "" {
		panic(result.Error)
	}
	var payloads commonpb.Payloads
	check(proto.Unmarshal(result.Payloads, &payloads))
	var value string
	check(converter.GetDefaultDataConverter().FromPayloads(&payloads, &value))
	if value != "done" {
		panic("unexpected workflow result: " + value)
	}
	completion.Reply(nil, nil)
	check(i.Wait())
	fmt.Println("sleep-for-days behavior passed:", os.Args[1])
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func marshal(value any) []byte {
	data, err := json.Marshal(value)
	check(err)
	return data
}

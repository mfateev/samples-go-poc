// Check each branch and failure path through the sample's real isolate entry.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"isolate"
	"reflect"
	"time"

	choice "github.com/mfateev/samples-go-poc/choice-exclusive"
	"github.com/mfateev/samples-go-poc/internal/testbridge"
	"github.com/mfateev/sdk-go-poc/workflow"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/proto"
)

func main() {
	handle, ok := isolate.LookupFunction(choice.ExclusiveChoice)
	if !ok || handle.Signature().NumIn() != 1 || handle.Signature().In(0) != reflect.TypeFor[context.Context]() || handle.Signature().NumOut() != 1 {
		panic("expected a marked context-only, error-only workflow")
	}
	for _, selected := range []string{"apple", "banana", "cherry", "orange"} {
		run(handle, selected, nil, nil, "")
	}
	run(handle, "", errors.New("get order failed"), nil, "get order failed")
	run(handle, "orange", nil, errors.New("ordering failed"), "ordering failed")
	run(handle, "pear", nil, nil, "unknown order choice: pear")
	fmt.Println("exclusive-choice branches and failures passed")
}

func run(handle isolate.Handle, selected string, getErr, orderErr error, wantErr string) {
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
	deadline := time.After(5 * time.Second)
	next := func(op uint32) *isolate.Command {
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
				if !ok || command.Op != op {
					panic(fmt.Sprintf("expected operation %d, got %+v", op, command))
				}
				return command
			case <-deadline:
				panic("workflow did not make progress")
			}

		}
	}
	start := next(workflow.OpStartPayloads)
	testbridge.ReplyStart(start, workflow.PayloadStart{Name: handle.Name()})
	activity := next(workflow.OpScheduleActivity)
	validateActivity(activity, "GetOrder", "")
	activity.Reply(nil, nil)
	await := next(workflow.OpAwaitActivity)
	await.Reply(activityOutcome(encodeResult(selected), getErr), nil)
	if getErr == nil && selected != "pear" {
		activity = next(workflow.OpScheduleActivity)
		name := map[string]string{"apple": "OrderApple", "banana": "OrderBanana", "cherry": "OrderCherry", "orange": "OrderOrange"}[selected]
		validateActivity(activity, name, selected)
		activity.Reply(nil, nil)
		await = next(workflow.OpAwaitActivity)
		await.Reply(activityOutcome(nil, orderErr), nil)
	}
	completion := next(workflow.OpCompletePayloads)
	var result workflow.PayloadCompletion
	check(json.Unmarshal(completion.Payload, &result))
	if result.Error != wantErr || len(result.Payloads) != 0 {
		panic(fmt.Sprintf("completion = %+v, expected error %q and no result", result, wantErr))
	}
	completion.Reply(nil, nil)
	check(i.Wait())
}

func validateActivity(command *isolate.Command, name, input string) {
	var request workflow.ActivityPayloadRequest
	check(json.Unmarshal(command.Payload, &request))
	var payloads commonpb.Payloads
	check(proto.Unmarshal(request.Payloads, &payloads))
	wantCount := 1
	if name == "GetOrder" {
		wantCount = 0
	}
	if len(payloads.Payloads) != wantCount {
		panic("wrong activity argument count")
	}
	var decoded string
	if wantCount != 0 {
		check(converter.GetDefaultDataConverter().FromPayloads(&payloads, &decoded))
	}
	if !request.Function || request.Name != name || decoded != input || request.Options == nil || request.Options.StartToCloseTimeout != 10*time.Second {
		panic(fmt.Sprintf("activity = %+v, expected %s(%q)", request, name, input))
	}

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

func encodeResult(value any) []byte {
	payloads, err := converter.GetDefaultDataConverter().ToPayloads(value)
	check(err)
	data, err := proto.Marshal(payloads)
	check(err)
	return data
}

func activityOutcome(payload []byte, err error) []byte {
	o := workflow.ActivityOutcome{Payloads: payload}
	if err != nil {
		o.Error = err.Error()
	}
	return marshal(o)
}

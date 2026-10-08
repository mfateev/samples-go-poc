// Drive the sample's real isolate calls without the Temporal task adapter.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"isolate"
	"os"
	"reflect"
	"runtime"
	"time"

	sleepfordays "github.com/mfateev/samples-go-poc/sleep-for-days"
	"github.com/mfateev/sdk-go-poc/workflow"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/proto"
)

func main() {
	scenario := os.Args[1]
	handle, ok := isolate.LookupFunction(sleepfordays.SleepForDays)
	if !ok || handle.Signature().NumIn() != 1 || handle.Signature().In(0) != reflect.TypeFor[context.Context]() {
		panic("expected a marked context-only workflow")
	}
	clock := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	i, err := isolate.New(isolate.Config{
		Deterministic: true,
		Program:       handle.ProgramWithHandle(func(entry isolate.Handle) { _ = workflow.RunFunction(entry) }),
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
	var timer, signal, cancellation *isolate.Command
	var ids []uint64
	waits := make(map[uint64]*isolate.Command)
	round := 0
	terminal := false
	var operations []uint32
	// After either half of an iteration completes, fence all native work and
	// prove no new email or timer was scheduled before the other half finished.
	checkWaiting := func() {
		suspended := make(chan error, 1)
		go func() { suspended <- i.Suspend() }()
		select {
		case err := <-suspended:
			check(err)
		case <-deadline:
			panic("workflow failed to suspend while waiting for iteration completion")
		}
		select {
		case command := <-i.Commands():
			panic(fmt.Sprintf("started next iteration too early: operation=%d", command.Op))
		default:
		}
	}
	for {
		var command *isolate.Command
		select {
		case command = <-i.Commands():
			if command == nil {
				panic("isolate stopped before completion")
			}
		case <-deadline:
			stack := make([]byte, 1<<20)
			n := runtime.Stack(stack, true)
			panic(fmt.Sprintf("workflow did not make progress: scenario=%s round=%d terminal=%v ids=%v operations=%v\n%s", scenario, round, terminal, ids, operations, stack[:n]))
		case <-i.Done():
			panic(fmt.Sprintf("isolate stopped before completion: %v", i.Wait()))
		}
		if command.Op == workflow.OpCancellableCall {
			var request workflow.CallRequest
			check(json.Unmarshal(command.Payload, &request))
			command.Op, command.Payload = request.Op, request.Payload
		}
		operations = append(operations, command.Op)
		switch command.Op {
		case workflow.OpStartPayloads:
			command.Reply(marshal(workflow.PayloadStart{Name: handle.Name()}), nil)
		case workflow.OpWorkflowCancel:
			cancellation = command
		case workflow.OpScheduleActivity:
			if terminal {
				panic("email scheduled after terminal event")
			}
			var request workflow.ActivityPayloadRequest
			check(json.Unmarshal(command.Payload, &request))
			var args commonpb.Payloads
			check(proto.Unmarshal(request.Payloads, &args))
			var message string
			check(converter.GetDefaultDataConverter().FromPayloads(&args, &message))
			if len(args.Payloads) != 1 || request.ID == 0 || request.Name != "SendEmail" || message != "Sleeping for 30 days" || request.Options == nil || request.Options.StartToCloseTimeout != 10*time.Second {
				panic("unexpected email request")
			}
			ids = append(ids, request.ID)
			command.Reply(nil, nil)
		case workflow.OpAwaitActivity:
			var id uint64
			check(json.Unmarshal(command.Payload, &id))
			if waits[id] != nil {
				panic("duplicate email result wait")
			}
			waits[id] = command
		case workflow.OpSleep:
			if terminal {
				panic("timer scheduled after terminal event")
			}
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
		case workflow.OpCancelCall, workflow.OpCancelActivity:
			command.Reply(nil, nil)
		case workflow.OpCompletePayloads:
			if !terminal {
				panic("workflow completed before any terminal event")
			}
			var result workflow.PayloadCompletion
			check(json.Unmarshal(command.Payload, &result))
			if scenario == "failed-email" || scenario == "late-failed-email" {
				if !result.Failed || result.Error != "email failed" || result.Canceled {
					panic(fmt.Sprintf("email error not propagated: %+v", result))
				}
			} else if scenario == "canceled" {
				if !result.Failed || !result.Canceled {
					panic("workflow cancellation lost")
				}
			} else {
				if result.Failed || result.Error != "" {
					panic("successful workflow failed: " + result.Error)
				}
				var payloads commonpb.Payloads
				check(proto.Unmarshal(result.Payloads, &payloads))
				var value string
				check(converter.GetDefaultDataConverter().FromPayloads(&payloads, &value))
				if value != "done" {
					panic("unexpected workflow result: " + value)
				}
			}
			command.Reply(nil, nil)
			check(i.Wait())
			fmt.Println("sleep-for-days behavior passed:", scenario)
			return
		default:
			panic(fmt.Sprintf("unexpected operation: %d", command.Op))
		}
		// Exercise each period only after all concurrent subscriptions arrive.
		if terminal || len(ids) != round+1 || waits[ids[round]] == nil || timer == nil || signal == nil || cancellation == nil {
			continue
		}
		switch {
		case scenario == "failed-email" || scenario == "late-failed-email":
			if scenario == "late-failed-email" {
				clock = clock.Add(30 * 24 * time.Hour)
				check(i.AdvanceTime(clock))
				timer.Reply(nil, nil)
				checkWaiting()
			}
			// An email failure still fails the workflow after its timer fires.
			waits[ids[0]].Reply(marshal(workflow.ActivityOutcome{Failed: true, Error: "email failed"}), nil)
			terminal = true
			if scenario == "late-failed-email" {
				check(i.Resume())
			}
		case scenario == "canceled":
			cancellation.Reply(nil, nil)
			terminal = true
		default:
			if scenario == "signal-first" {
				signal.Reply(marshal(workflow.Signal{Name: "complete"}), nil)
				signal = nil
				terminal = true
				continue
			}
			if scenario == "email-first" {
				waits[ids[round]].Reply(marshal(workflow.ActivityOutcome{}), nil)
				checkWaiting()
			} else {
				clock = clock.Add(30 * 24 * time.Hour)
				check(i.AdvanceTime(clock))
				timer.Reply(nil, nil)
				checkWaiting()
			}
			if scenario == "pending-email" || round == 2 {
				// Finish with either the email or the timer still pending.
				signal.Reply(marshal(workflow.Signal{Name: "complete"}), nil)
				signal = nil
				terminal = true
			} else {
				if scenario == "email-first" {
					clock = clock.Add(30 * 24 * time.Hour)
					check(i.AdvanceTime(clock))
					timer.Reply(nil, nil)
				} else {
					waits[ids[round]].Reply(marshal(workflow.ActivityOutcome{}), nil)
				}
				timer = nil
				round++
			}
			check(i.Resume())
		}
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

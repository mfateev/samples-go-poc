package updatabletimer

import (
	"time"

	"context"

	"github.com/mfateev/sdk-go-poc/workflow"
)

const (
	TaskQueue  = "updatable-timer"
	WorkflowID = "updatable-timer"
	QueryType  = "GetWakeUpTime"
	SignalType = "UpdateWakeUpTime"
)

// UpdatableTimer is an example of a timer that can have its wake time updated
type UpdatableTimer struct {
	wakeUpTime time.Time
}

// SleepUntil sleeps until the provided wake-up time.
// The wake-up time can be updated at any time by sending a new time over updateWakeUpTimeCh.
// Supports ctx cancellation.
// Returns context.Canceled if ctx was canceled.
func (u *UpdatableTimer) SleepUntil(ctx context.Context, wakeUpTime time.Time, updates <-chan time.Time) error {
	u.wakeUpTime = wakeUpTime
	timer := time.NewTimer(time.Until(wakeUpTime))
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return nil
		case wake, ok := <-updates:
			if !ok {
				return ctx.Err()
			}
			u.wakeUpTime = wake
			timer.Reset(time.Until(wake))
		}
	}
}

func (u *UpdatableTimer) GetWakeUpTime() time.Time {
	return u.wakeUpTime
}

// Workflow that sleeps initialWakeUpTime unless the new wake-up time is received through "UpdateWakeUpTime" signal.
//
//temporal:workflow
func Workflow(ctx context.Context, initialWakeUpTime time.Time) error {
	timer := UpdatableTimer{}
	err := workflow.SetQueryHandler(ctx, QueryType, func() (time.Time, error) {
		return timer.GetWakeUpTime(), nil
	})
	if err != nil {
		return err
	}
	return timer.SleepUntil(ctx, initialWakeUpTime, workflow.GetSignalChannel[time.Time](ctx, SignalType))
}

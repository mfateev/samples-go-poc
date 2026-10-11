package infrequent

import (
	"context"

	"github.com/mfateev/samples-go-poc/polling"
)

type PollingActivities struct {
	TestService *polling.TestService
}

// DoPoll Activity.
func (a *PollingActivities) DoPoll(cmd context.Context) (string, error) {
	return a.TestService.GetServiceResult(cmd)
}

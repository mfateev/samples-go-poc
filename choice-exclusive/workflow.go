// The upstream exclusive-choice workflow, adapted to run in an isolate.
package choiceexclusive

import (
	"errors"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
)

//go:isolate
func ExclusiveChoice() (string, error) {
	choice, err := workflow.ExecuteActivity("GetOrder", nil, 10*time.Second)
	if err != nil {
		return "", err
	}
	var activityName string
	switch string(choice) {
	case "apple":
		activityName = "OrderApple"
	case "banana":
		activityName = "OrderBanana"
	case "cherry":
		activityName = "OrderCherry"
	case "orange":
		activityName = "OrderOrange"
	default:
		return "", errors.New("unknown order choice: " + string(choice))
	}
	_, err = workflow.ExecuteActivity(activityName, choice, 10*time.Second)
	return string(choice), err
}

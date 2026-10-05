// The upstream exclusive-choice workflow, adapted to run in an isolate.
package choiceexclusive

import (
	"fmt"
	"time"

	"github.com/mfateev/sdk-go-poc/workflow"
)

const (
	OrderChoiceApple  = "apple"
	OrderChoiceBanana = "banana"
	OrderChoiceCherry = "cherry"
	OrderChoiceOrange = "orange"
)

//go:isolate
func ExclusiveChoice() error {
	choice, err := workflow.ExecuteActivity("GetOrder", nil, 10*time.Second)
	if err != nil {
		return err
	}
	var activityName string
	switch string(choice) {
	case OrderChoiceApple:
		activityName = "OrderApple"
	case OrderChoiceBanana:
		activityName = "OrderBanana"
	case OrderChoiceCherry:
		activityName = "OrderCherry"
	case OrderChoiceOrange:
		activityName = "OrderOrange"
	default:
		return fmt.Errorf("unknown order choice: %v", string(choice))
	}
	_, err = workflow.ExecuteActivity(activityName, choice, 10*time.Second)
	return err
}

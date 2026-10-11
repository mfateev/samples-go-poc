package branch

import (
	"context"
	"fmt"
)

// @@@SNIPSTART samples-go-branch-activity-definition
// SampleActivity is a Temporal Activity Definition
func SampleActivity(ctx context.Context, input string) (string, error) {
	name := "sampleActivity"
	fmt.Printf("Run %s with input %v \n", name, input)
	return "Result_" + input, nil
}

// @@@SNIPEND

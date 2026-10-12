package random_streams

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/mfateev/sdk-go-poc/testsuite"
	"github.com/stretchr/testify/require"
)

func TestNamedStreamIndependenceAndCompletedQueries(t *testing.T) {
	for _, procs := range []int{1, 2, 8} {
		t.Run(fmt.Sprint(procs), func(t *testing.T) {
			previous := runtime.GOMAXPROCS(procs)
			defer runtime.GOMAXPROCS(previous)
			var baseline Result
			for _, audit := range []int{0, 123} {
				var suite testsuite.WorkflowTestSuite
				env := suite.NewTestWorkflowEnvironment(t)
				env.ExecuteWorkflow(Plan, Input{Count: 10, AuditDraws: audit})
				require.NoError(t, env.GetWorkflowError())
				var result Result
				require.NoError(t, env.GetWorkflowResult(&result))
				if audit == 0 {
					baseline = result
				} else {
					require.Equal(t, baseline, result, "audit draws changed another stream")
				}
				var first int
				for i := range 3 {
					query, err := env.QueryWorkflow("next")
					require.NoError(t, err)
					var next int
					require.NoError(t, query.Get(&next))
					require.GreaterOrEqual(t, next, 1)
					require.LessOrEqual(t, next, 6)
					if i == 0 {
						first = next
					} else {
						require.Equal(t, first, next, "query advanced workflow stream")
					}
				}
			}
		})
	}
}
func TestInvalidRandomStreamInput(t *testing.T) {
	for _, input := range []Input{{Count: 0}, {Count: 101}, {Count: 1, AuditDraws: -1}} {
		var suite testsuite.WorkflowTestSuite
		env := suite.NewTestWorkflowEnvironment(t)
		env.ExecuteWorkflow(Plan, input)
		require.ErrorContains(t, env.GetWorkflowError(), "count must be")
	}
}

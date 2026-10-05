package sleepfordays

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// Test the actual marked function with a host that explicitly services every
// concurrent Call. This verifies the sample independently of the unfinished
// Temporal bridge quiescence barrier.
func TestSleepForDaysBehavior(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "sleep-for-days-driver")
	goTool := filepath.Join(runtime.GOROOT(), "bin", "go")
	build := exec.Command(goTool, "build", "-o", binary, "./testdata/driver")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build driver: %v\n%s", err, output)
	}
	for _, scenario := range []string{"pending-email", "failed-email"} {
		t.Run(scenario, func(t *testing.T) {
			if output, err := exec.Command(binary, scenario).CombinedOutput(); err != nil {
				t.Fatalf("driver: %v\n%s", err, output)
			}
		})
	}
}

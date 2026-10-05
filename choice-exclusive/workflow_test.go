package choiceexclusive

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestExclusiveChoiceBehavior(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "choice-driver")
	build := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, "./testdata/driver")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build driver: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("driver: %v\n%s", err, output)
	}
}

func TestConfiguredOrderActivities(t *testing.T) {
	orders := &OrderActivities{OrderChoices: []string{OrderChoiceCherry}}
	selected, err := orders.GetOrder()
	if err != nil || selected != OrderChoiceCherry {
		t.Fatalf("configured order = %q, %v", selected, err)
	}
}

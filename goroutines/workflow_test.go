package goroutines

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGreetAllNativeGoroutines(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "goroutines-driver")
	build := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, "./testdata/driver")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("driver: %v\n%s", err, output)
	}
}

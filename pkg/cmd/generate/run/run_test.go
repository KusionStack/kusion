package run

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKPMRunnerRun(t *testing.T) {
	currentPath, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	workDir := filepath.Join(currentPath, "testdata/prod")
	codeRunner := &KPMRunner{}
	value, err := codeRunner.Run(workDir, nil)
	if err != nil {
		t.Fatalf("Failed to run configuration code: %v", err)
	}
	if len(value) == 0 {
		t.Fatalf("Unexpected value output")
	}
}

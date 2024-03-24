package run

import (
	"fmt"
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
	fmt.Println(workDir)
	/*codeRunner := &KPMRunner{}
	arguments := make(map[string]string, 1)
	arguments["include_schema_type_path"] = "true"
	value, err := codeRunner.Run(workDir, arguments)
	if err != nil {
		t.Fatalf("Failed to run configuration code: %v", err)
	}
	if len(value) == 0 {
		t.Fatalf("Unexpected value output")
	}*/
}

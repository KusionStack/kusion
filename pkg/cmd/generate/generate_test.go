package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	currentPath, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	workDir := filepath.Join(currentPath, "run/testdata/prod")
	fmt.Println(workDir)
	/*gtr := &generator.DefaultGenerator{Runner: &run.KPMRunner{}}
	opt := &GenerateOptions{
		WorkDir:   workDir,
		Generator: gtr,
	}
	err = opt.Run()
	if err != nil {
		t.Fatalf("Failed to generate spec as %v", err)
	}*/
}

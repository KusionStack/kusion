package fake

import "kusionstack.io/kusion/pkg/engine/api/generate/run"

var _ run.CodeRunner = &KPMRunner{}

// KPMRunner is a fake code runner for testing purposes.
type KPMRunner struct{}

// Run does nothing.
func (r *KPMRunner) Run(workDir string, arguments map[string]string) ([]byte, error) {
	return nil, nil
}

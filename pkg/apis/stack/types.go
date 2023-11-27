package stack

import (
	"errors"

	"github.com/pterm/pterm"

	"kusionstack.io/kusion/pkg/log"
)

const File = "stack.yaml"

var (
	ErrNotStackDirectory = errors.New("path must be a stack directory")
	ErrStackNotUnique    = errors.New("the stack obtained is not unique")
)

// Configuration is the stack configuration
type Configuration struct {
	Name       string `json:"name" yaml:"name"`             // Stack name
	KubeConfig string `json:"kubeConfig" yaml:"kubeConfig"` // KubeConfig file path for this stack
}

type Stack struct {
	Configuration `json:",inline" yaml:",inline"`
	Path          string `json:"path,omitempty" yaml:"path,omitempty"` // Absolute path to the stack directory
}

// NewStack creates a new stack
func NewStack(config *Configuration, path string) *Stack {
	return &Stack{
		Configuration: *config,
		Path:          path,
	}
}

// GetName returns the name of the stack
func (s *Stack) GetName() string {
	return s.Name
}

// GetPath returns the path of the stack
func (s *Stack) GetPath() string {
	return s.Path
}

// TableReport returns the report string of table format
func (s *Stack) TableReport() string {
	// Fill table header
	tableHeader := []string{"Type", "Name"}
	tableData := pterm.TableData{tableHeader}

	// Fill table content
	tableData = append(tableData, []string{"Stack Name", s.GetName()})
	if s.GetPath() != "" {
		tableData = append(tableData, []string{"Stack Path", s.GetPath()})
	}

	// Render table
	report, err := pterm.DefaultTable.WithHasHeader().
		WithBoxed(true).
		WithData(tableData).
		Srender()
	if err != nil {
		log.Warn(err)
	}

	return report
}

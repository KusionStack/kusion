package projectstack

import (
	"errors"
	"strings"

	"github.com/pterm/pterm"

	"kusionstack.io/kusion/pkg/engine/backend"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/vals"
)

var (
	ErrNotStackDirectory   = errors.New("path must be a stack directory")
	ErrNotProjectDirectory = errors.New("path must be a project directory")
	ErrProjectNotUnique    = errors.New("the project obtained is not unique")
	ErrStackNotUnique      = errors.New("the stack obtained is not unique")
)

const (
	StackFile                               = "stack.yaml"
	ProjectFile                             = "project.yaml"
	CiTestDir                               = "ci-test"
	SettingsFile                            = "settings.yaml"
	StdoutGoldenFile                        = "stdout.golden.yaml"
	KclFile                                 = "kcl.yaml"
	KCLGenerator              GeneratorType = "KCL"
	AppConfigurationGenerator GeneratorType = "AppConfiguration"
)

type GeneratorType string

// GeneratorConfig represent Generator configs saved in project.yaml
type GeneratorConfig struct {
	Type    GeneratorType          `json:"type"`
	Configs map[string]interface{} `json:"configs,omitempty"`
}

// ProjectConfiguration is the project configuration
type ProjectConfiguration struct {
	// Project name
	Name string `json:"name" yaml:"name"`

	// Tenant name
	Tenant string `json:"tenant,omitempty" yaml:"tenant,omitempty"`

	// Backend storage config
	Backend *backend.Storage `json:"backend,omitempty" yaml:"backend,omitempty"`

	// SpecGenerator configs
	Generator *GeneratorConfig `json:"generator,omitempty" yaml:"generator,omitempty"`

	// Secret stores
	SecretStores *vals.SecretStores `json:"secret_stores,omitempty" yaml:"secret_stores,omitempty"`
}

type Project struct {
	ProjectConfiguration `json:",inline" yaml:",inline"`
	Path                 string   `json:"path,omitempty" yaml:"path,omitempty"`     // Absolute path to the project directory
	Stacks               []*Stack `json:"stacks,omitempty" yaml:"stacks,omitempty"` // Stacks
}

// NewProject creates a new project
func NewProject(config *ProjectConfiguration, path string, stacks []*Stack) *Project {
	return &Project{
		ProjectConfiguration: *config,
		Path:                 path,
		Stacks:               stacks,
	}
}

// GetName returns the name of the project
func (p *Project) GetName() string {
	return p.Name
}

// GetName returns the path of the project
func (p *Project) GetPath() string {
	return p.Path
}

// TableReport returns the report string of table format
func (p *Project) TableReport() string {
	// Fill table header
	tableHeader := []string{"Type", "Name"}
	tableData := pterm.TableData{tableHeader}

	// Fill table content
	tableData = append(tableData, []string{"Project Name", p.GetName()})
	if p.GetPath() != "" {
		tableData = append(tableData, []string{"Project Path", p.GetPath()})
	}

	if p.Tenant != "" {
		tableData = append(tableData, []string{"Tenant", p.Tenant})
	}

	stacksList := []string{}
	for _, s := range p.Stacks {
		stacksList = append(stacksList, s.GetName())
	}

	tableData = append(tableData, []string{"Stacks", strings.Join(stacksList, ",")})

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

// StackConfiguration is the stack configuration
type StackConfiguration struct {
	Name string `json:"name" yaml:"name"` // Stack name
}

type Stack struct {
	StackConfiguration `json:",inline" yaml:",inline"`
	Path               string `json:"path,omitempty" yaml:"path,omitempty"` // Absolute path to the stack directory
}

// NewStack creates a new stack
func NewStack(config *StackConfiguration, path string) *Stack {
	return &Stack{
		StackConfiguration: *config,
		Path:               path,
	}
}

// GetName returns the name of the stack
func (s *Stack) GetName() string {
	return s.Name
}

// GetName returns the path of the stack
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

package project

import (
	"errors"
	"strings"

	"github.com/pterm/pterm"

	"kusionstack.io/kusion/pkg/apis/stack"
	"kusionstack.io/kusion/pkg/engine/backend"
	"kusionstack.io/kusion/pkg/log"
)

var (
	ErrNotProjectDirectory = errors.New("path must be a project directory")
	ErrProjectNotUnique    = errors.New("the project obtained is not unique")
)

const (
	ProjectFile                         = "project.yaml"
	KclFile                             = "kcl.yaml"
	KCLBuilder              BuilderType = "KCL"
	AppConfigurationBuilder BuilderType = "AppConfiguration"
	PodMonitorType          MonitorType = "Pod"
	ServiceMonitorType      MonitorType = "Service"
)

type (
	BuilderType string
	MonitorType string
)

// GeneratorConfig represent Generator configs saved in project.yaml
type GeneratorConfig struct {
	Type    BuilderType            `json:"type"`
	Configs map[string]interface{} `json:"configs,omitempty"`
}

// PrometheusConfig represent Prometheus configs saved in project.yaml
type PrometheusConfig struct {
	OperatorMode bool        `yaml:"operatorMode,omitempty" json:"operatorMode,omitempty"`
	MonitorType  MonitorType `yaml:"monitorType,omitempty" json:"monitorType,omitempty"`
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

	// Prometheus configs
	Prometheus *PrometheusConfig `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`
}

type Project struct {
	ProjectConfiguration `json:",inline" yaml:",inline"`
	Path                 string         `json:"path,omitempty" yaml:"path,omitempty"`     // Absolute path to the project directory
	Stacks               []*stack.Stack `json:"stacks,omitempty" yaml:"stacks,omitempty"` // Stacks
}

// NewProject creates a new project
func NewProject(config *ProjectConfiguration, path string, stacks []*stack.Stack) *Project {
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

// GetPath returns the path of the project
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

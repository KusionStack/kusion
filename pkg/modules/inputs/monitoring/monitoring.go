package monitoring

import (
	"errors"

	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

const (
	ModuleName                     = "monitoring"
	OperatorModeKey                = "operatorMode"
	MonitorTypeKey                 = "monitorType"
	IntervalKey                    = "interval"
	TimeoutKey                     = "timeout"
	SchemeKey                      = "scheme"
	defaultInterval                = "15s"
	defaultTimeout                 = "30s"
	defaultScheme                  = "http"
	PodMonitorType     MonitorType = "Pod"
	ServiceMonitorType MonitorType = "Service"
)

var (
	ErrTimeoutGreaterThanInterval = errors.New("timeout cannot be greater than interval")
)

type (
	MonitorType string
)

type Monitor struct {
	OperatorMode bool                  `yaml:"operatorMode,omitempty" json:"operatorMode,omitempty"`
	Interval     prometheusv1.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout      prometheusv1.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	MonitorType  MonitorType           `yaml:"monitorType,omitempty" json:"monitorType,omitempty"`
	Path         string                `yaml:"path,omitempty" json:"path,omitempty"`
	// Despite what the name suggests, PodMonitor and ServiceMonitor actually
	// only accept port names as the input. So in operator mode, this port field
	// need to be the user-provided port name.
	Port   string `yaml:"port,omitempty" json:"port,omitempty"`
	Scheme string `yaml:"scheme,omitempty" json:"scheme,omitempty"`
}

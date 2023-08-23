package monitoring

import (
	prometheusV1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

type Monitor struct {
	Interval     prometheusV1.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout      prometheusV1.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Path         string                `yaml:"path,omitempty" json:"path,omitempty"`
	Port         string                `yaml:"port,omitempty" json:"port,omitempty"`
	Scheme       string                `yaml:"scheme,omitempty" json:"scheme,omitempty"`
	OperatorMode bool                  `yaml:"operatorMode,omitempty" json:"operatorMode,omitempty"`
	MonitorType  string                `yaml:"monitorType,omitempty" json:"monitorType,omitempty"`
}

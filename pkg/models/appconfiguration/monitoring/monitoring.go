package monitoring

import (
	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

type Monitor struct {
	Interval prometheusv1.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout  prometheusv1.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Path     string                `yaml:"path,omitempty" json:"path,omitempty"`
	// Despite what the name suggests, PodMonitor and ServiceMonitor actually
	// only accept port names as the input. So in operator mode, this port field
	// need to be the user-provided port name.
	Port   string `yaml:"port,omitempty" json:"port,omitempty"`
	Scheme string `yaml:"scheme,omitempty" json:"scheme,omitempty"`
}

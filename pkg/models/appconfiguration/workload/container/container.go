package container

import (
	"encoding/json"
	"errors"
)

type Container struct {
	// Image to run for this container
	Image string `yaml:"image" json:"image"`
	// Entrypoint array.
	// The image's ENTRYPOINT is used if this is not provided.
	Command []string `yaml:"command,omitempty" json:"command,omitempty"`
	// Arguments to the entrypoint.
	// The image's CMD is used if this is not provided.
	Args []string `yaml:"args,omitempty" json:"args,omitempty"`
	// Collection of environment variables to set in the container.
	// The value of environment variable may be static text or a value from a secret.
	Env map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	// The current working directory of the running process defined in entrypoint.
	WorkingDir string `yaml:"workingDir,omitempty" json:"workingDir,omitempty"`
	// Periodic probe of container liveness.
	LivenessProbe *Probe `yaml:"livenessProbe,omitempty" json:"livenessProbe,omitempty"`
	// Periodic probe of container service readiness.
	ReadinessProbe *Probe `yaml:"readinessProbe,omitempty" json:"readinessProbe,omitempty"`
	// StartupProbe indicates that the Pod has successfully initialized.
	StartupProbe *Probe `yaml:"startupProbe,omitempty" json:"startupProbe,omitempty"`
}

// Probe describes a health check to be performed against a container to determine whether it is
// alive or ready to receive traffic.
type Probe struct {
	// The action taken to determine the health of a container.
	ProbeHandler *ProbeHandler `yaml:"probeHandler" json:"probeHandler"`
	// Number of seconds after the container has started before liveness probes are initiated.
	InitialDelaySeconds int32 `yaml:"initialDelaySeconds,omitempty" json:"initialDelaySeconds,omitempty"`
	// Number of seconds after which the probe times out.
	TimeoutSeconds int32 `yaml:"timeoutSeconds,omitempty" json:"timeoutSeconds,omitempty"`
	// How often (in seconds) to perform the probe.
	PeriodSeconds int32 `yaml:"periodSeconds,omitempty" json:"periodSeconds,omitempty"`
	// Minimum consecutive successes for the probe to be considered successful after having failed.
	SuccessThreshold int32 `yaml:"successThreshold,omitempty" json:"successThreshold,omitempty"`
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	FailureThreshold int32 `yaml:"failureThreshold,omitempty" json:"failureThreshold,omitempty"`
}

type ProbeType struct {
	// Type of action to be taken.
	Type string `yaml:"_type" json:"_type"`
}

// ProbeHandler defines a specific action that should be taken in a probe.
// One and only one of the fields must be specified.
type ProbeHandler struct {
	// Type of action to be taken.
	ProbeType `yaml:"_type" json:"_type"`
	// Exec specifies the action to take.
	// +optional
	*ExecAction `yaml:",inline" json:",inline"`
	// HTTPGet specifies the http request to perform.
	// +optional
	*HTTPGetAction `yaml:",inline" json:",inline"`
	// TCPSocket specifies an action involving a TCP port.
	// +optional
	*TCPSocketAction `yaml:",inline" json:",inline"`
}

// ExecAction describes a "run in container" action.
type ExecAction struct {
	// Command is the command line to execute inside the container, the working directory for the
	// command  is root ('/') in the container's filesystem.
	// Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
	Command []string `yaml:"command,omitempty" json:"command,omitempty"`
}

// HTTPGetAction describes an action based on HTTP Get requests.
type HTTPGetAction struct {
	// URL is the full qualified url location to send HTTP requests.
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
	// Custom headers to set in the request. HTTP allows repeated headers.
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
}

// TCPSocketAction describes an action based on opening a socket.
type TCPSocketAction struct {
	// URL is the full qualified url location to open a socket.
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
}

// MarshalJSON implements the json.Marshaler interface for ProbeHandler.
func (p *ProbeHandler) MarshalJSON() ([]byte, error) {
	switch p.Type {
	case "Http":
		return json.Marshal(struct {
			ProbeType      `json:",inline"`
			*HTTPGetAction `json:",inline"`
		}{
			ProbeType:     ProbeType{p.Type},
			HTTPGetAction: p.HTTPGetAction,
		})
	case "Exec":
		return json.Marshal(struct {
			ProbeType   `json:",inline"`
			*ExecAction `json:",inline"`
		}{
			ProbeType:  ProbeType{p.Type},
			ExecAction: p.ExecAction,
		})
	case "Tcp":
		return json.Marshal(struct {
			ProbeType        `json:",inline"`
			*TCPSocketAction `json:",inline"`
		}{
			ProbeType:       ProbeType{p.Type},
			TCPSocketAction: p.TCPSocketAction,
		})
	default:
		return nil, errors.New("unrecognized probe handler type")
	}
}

// UnmarshalJSON implements the json.Unmarshaller interface for ProbeHandler.
func (p *ProbeHandler) UnmarshalJSON(data []byte) error {
	var probeType ProbeType
	err := json.Unmarshal(data, &probeType)
	if err != nil {
		return err
	}

	p.Type = probeType.Type
	switch p.Type {
	case "Http":
		handler := &HTTPGetAction{}
		err = json.Unmarshal(data, handler)
		p.HTTPGetAction = handler
	case "Exec":
		handler := &ExecAction{}
		err = json.Unmarshal(data, handler)
		p.ExecAction = handler
	case "Tcp":
		handler := &TCPSocketAction{}
		err = json.Unmarshal(data, handler)
		p.TCPSocketAction = handler
	default:
		return errors.New("unrecognized probe handler type")
	}

	return err
}

// MarshalYAML implements the yaml.Marshaler interface for ProbeHandler.
func (p *ProbeHandler) MarshalYAML() (interface{}, error) {
	switch p.Type {
	case "Http":
		return struct {
			ProbeType     `yaml:",inline" json:",inline"`
			HTTPGetAction `yaml:",inline" json:",inline"`
		}{
			ProbeType:     ProbeType{Type: "Http"},
			HTTPGetAction: *p.HTTPGetAction,
		}, nil
	case "Exec":
		return struct {
			ProbeType  `yaml:",inline" json:",inline"`
			ExecAction `yaml:",inline" json:",inline"`
		}{
			ProbeType:  ProbeType{Type: "Exec"},
			ExecAction: *p.ExecAction,
		}, nil
	case "Tcp":
		return struct {
			ProbeType       `yaml:",inline" json:",inline"`
			TCPSocketAction `yaml:",inline" json:",inline"`
		}{
			ProbeType:       ProbeType{Type: "Tcp"},
			TCPSocketAction: *p.TCPSocketAction,
		}, nil
	}

	return nil, nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for ProbeHandler.
func (p *ProbeHandler) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var probeType ProbeType
	err := unmarshal(&probeType)
	if err != nil {
		return err
	}

	p.Type = probeType.Type
	switch p.Type {
	case "Http":
		handler := &HTTPGetAction{}
		err = unmarshal(handler)
		p.HTTPGetAction = handler
	case "Exec":
		handler := &ExecAction{}
		err = unmarshal(handler)
		p.ExecAction = handler
	case "Tcp":
		handler := &TCPSocketAction{}
		err = unmarshal(handler)
		p.TCPSocketAction = handler
	default:
		return errors.New("unrecognized probe handler type")
	}

	return err
}

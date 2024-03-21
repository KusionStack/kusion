package container

import (
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/yaml.v2"

	"kusionstack.io/kusion/pkg/apis/core"
)

const (
	ProbePrefix = "v1.workload.container.probe."
	TypeHTTP    = core.BuiltinModulePrefix + ProbePrefix + "Http"
	TypeExec    = core.BuiltinModulePrefix + ProbePrefix + "Exec"
	TypeTCP     = core.BuiltinModulePrefix + ProbePrefix + "Tcp"
)

// Container describes how the App's tasks are expected to be run.
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
	Env yaml.MapSlice `yaml:"env,omitempty" json:"env,omitempty"`
	// The current working directory of the running process defined in entrypoint.
	WorkingDir string `yaml:"workingDir,omitempty" json:"workingDir,omitempty"`
	// Resource requirements for this container.
	Resources map[string]string `yaml:"resources,omitempty" json:"resources,omitempty"`
	// Files configures one or more files to be created in the container.
	Files map[string]FileSpec `yaml:"files,omitempty" json:"files,omitempty"`
	// Dirs configures one or more volumes to be mounted to the specified folder.
	Dirs map[string]string `yaml:"dirs,omitempty" json:"dirs,omitempty"`
	// Periodic probe of container liveness.
	LivenessProbe *Probe `yaml:"livenessProbe,omitempty" json:"livenessProbe,omitempty"`
	// Periodic probe of container service readiness.
	ReadinessProbe *Probe `yaml:"readinessProbe,omitempty" json:"readinessProbe,omitempty"`
	// StartupProbe indicates that the Pod has successfully initialized.
	StartupProbe *Probe `yaml:"startupProbe,omitempty" json:"startupProbe,omitempty"`
	// Actions that the management system should take in response to container lifecycle events.
	Lifecycle *Lifecycle `yaml:"lifecycle,omitempty" json:"lifecycle,omitempty"`
}

// FileSpec defines the target file in a Container
type FileSpec struct {
	// The content of target file in plain text.
	Content string `yaml:"content,omitempty" json:"content,omitempty"`
	// Source for the file content, might be a reference to a secret value.
	ContentFrom string `yaml:"contentFrom,omitempty" json:"contentFrom,omitempty"`
	// Mode bits used to set permissions on this file.
	Mode string `yaml:"mode" json:"mode"`
}

// TypeWrapper is a thin wrapper to make YAML decoder happy.
type TypeWrapper struct {
	// Type of action to be taken.
	Type string `yaml:"_type" json:"_type"`
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

// ProbeHandler defines a specific action that should be taken in a probe.
// One and only one of the fields must be specified.
type ProbeHandler struct {
	// Type of action to be taken.
	TypeWrapper `yaml:"_type" json:"_type"`
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

// Lifecycle describes actions that the management system should take in response
// to container lifecycle events.
type Lifecycle struct {
	// PreStop is called immediately before a container is terminated due to an
	// API request or management event such as liveness/startup probe failure,
	// preemption, resource contention, etc.
	PreStop *LifecycleHandler `yaml:"preStop,omitempty" json:"preStop,omitempty"`
	// PostStart is called immediately after a container is created.
	PostStart *LifecycleHandler `yaml:"postStart,omitempty" json:"postStart,omitempty"`
}

// LifecycleHandler defines a specific action that should be taken in a lifecycle
// hook. One and only one of the fields, except TCPSocket must be specified.
type LifecycleHandler struct {
	// Type of action to be taken.
	TypeWrapper `yaml:"_type" json:"_type"`
	// Exec specifies the action to take.
	// +optional
	*ExecAction `yaml:",inline" json:",inline"`
	// HTTPGet specifies the http request to perform.
	// +optional
	*HTTPGetAction `yaml:",inline" json:",inline"`
}

// MarshalJSON implements the json.Marshaler interface for ProbeHandler.
func (p *ProbeHandler) MarshalJSON() ([]byte, error) {
	switch p.Type {
	case TypeHTTP:
		return json.Marshal(struct {
			TypeWrapper    `json:",inline"`
			*HTTPGetAction `json:",inline"`
		}{
			TypeWrapper:   TypeWrapper{p.Type},
			HTTPGetAction: p.HTTPGetAction,
		})
	case TypeExec:
		return json.Marshal(struct {
			TypeWrapper `json:",inline"`
			*ExecAction `json:",inline"`
		}{
			TypeWrapper: TypeWrapper{p.Type},
			ExecAction:  p.ExecAction,
		})
	case TypeTCP:
		return json.Marshal(struct {
			TypeWrapper      `json:",inline"`
			*TCPSocketAction `json:",inline"`
		}{
			TypeWrapper:     TypeWrapper{p.Type},
			TCPSocketAction: p.TCPSocketAction,
		})
	default:
		return nil, fmt.Errorf("unrecognized probe handler type: %s", p.Type)
	}
}

// UnmarshalJSON implements the json.Unmarshaller interface for ProbeHandler.
func (p *ProbeHandler) UnmarshalJSON(data []byte) error {
	var probeType TypeWrapper
	err := json.Unmarshal(data, &probeType)
	if err != nil {
		return err
	}

	p.Type = probeType.Type
	switch p.Type {
	case TypeHTTP:
		handler := &HTTPGetAction{}
		err = json.Unmarshal(data, handler)
		p.HTTPGetAction = handler
	case TypeExec:
		handler := &ExecAction{}
		err = json.Unmarshal(data, handler)
		p.ExecAction = handler
	case TypeTCP:
		handler := &TCPSocketAction{}
		err = json.Unmarshal(data, handler)
		p.TCPSocketAction = handler
	default:
		return fmt.Errorf("unrecognized probe handler type: %s", p.Type)
	}

	return err
}

// MarshalYAML implements the yaml.Marshaler interface for ProbeHandler.
func (p *ProbeHandler) MarshalYAML() (interface{}, error) {
	switch p.Type {
	case TypeHTTP:
		return struct {
			TypeWrapper   `yaml:",inline" json:",inline"`
			HTTPGetAction `yaml:",inline" json:",inline"`
		}{
			TypeWrapper:   TypeWrapper{Type: p.Type},
			HTTPGetAction: *p.HTTPGetAction,
		}, nil
	case TypeExec:
		return struct {
			TypeWrapper `yaml:",inline" json:",inline"`
			ExecAction  `yaml:",inline" json:",inline"`
		}{
			TypeWrapper: TypeWrapper{Type: p.Type},
			ExecAction:  *p.ExecAction,
		}, nil
	case TypeTCP:
		return struct {
			TypeWrapper     `yaml:",inline" json:",inline"`
			TCPSocketAction `yaml:",inline" json:",inline"`
		}{
			TypeWrapper:     TypeWrapper{Type: p.Type},
			TCPSocketAction: *p.TCPSocketAction,
		}, nil
	}

	return nil, nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for ProbeHandler.
func (p *ProbeHandler) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var probeType TypeWrapper
	err := unmarshal(&probeType)
	if err != nil {
		return err
	}

	p.Type = probeType.Type
	switch p.Type {
	case TypeHTTP:
		handler := &HTTPGetAction{}
		err = unmarshal(handler)
		p.HTTPGetAction = handler
	case TypeExec:
		handler := &ExecAction{}
		err = unmarshal(handler)
		p.ExecAction = handler
	case TypeTCP:
		handler := &TCPSocketAction{}
		err = unmarshal(handler)
		p.TCPSocketAction = handler
	default:
		return fmt.Errorf("unrecognized probe handler type: %s", p.Type)
	}

	return err
}

// MarshalJSON implements the json.Marshaler interface for LifecycleHandler.
func (l *LifecycleHandler) MarshalJSON() ([]byte, error) {
	switch l.Type {
	case TypeHTTP:
		return json.Marshal(struct {
			TypeWrapper    `json:",inline"`
			*HTTPGetAction `json:",inline"`
		}{
			TypeWrapper:   TypeWrapper{l.Type},
			HTTPGetAction: l.HTTPGetAction,
		})
	case TypeExec:
		return json.Marshal(struct {
			TypeWrapper `json:",inline"`
			*ExecAction `json:",inline"`
		}{
			TypeWrapper: TypeWrapper{l.Type},
			ExecAction:  l.ExecAction,
		})
	default:
		return nil, errors.New("unrecognized lifecycle handler type")
	}
}

// UnmarshalJSON implements the json.Unmarshaller interface for LifecycleHandler.
func (l *LifecycleHandler) UnmarshalJSON(data []byte) error {
	var handlerType TypeWrapper
	err := json.Unmarshal(data, &handlerType)
	if err != nil {
		return err
	}

	l.Type = handlerType.Type
	switch l.Type {
	case TypeHTTP:
		handler := &HTTPGetAction{}
		err = json.Unmarshal(data, handler)
		l.HTTPGetAction = handler
	case TypeExec:
		handler := &ExecAction{}
		err = json.Unmarshal(data, handler)
		l.ExecAction = handler
	default:
		return errors.New("unrecognized lifecycle handler type")
	}

	return err
}

// MarshalYAML implements the yaml.Marshaler interface for LifecycleHandler.
func (l *LifecycleHandler) MarshalYAML() (interface{}, error) {
	switch l.Type {
	case TypeHTTP:
		return struct {
			TypeWrapper   `yaml:",inline" json:",inline"`
			HTTPGetAction `yaml:",inline" json:",inline"`
		}{
			TypeWrapper:   TypeWrapper{Type: l.Type},
			HTTPGetAction: *l.HTTPGetAction,
		}, nil
	case TypeExec:
		return struct {
			TypeWrapper `yaml:",inline" json:",inline"`
			ExecAction  `yaml:",inline" json:",inline"`
		}{
			TypeWrapper: TypeWrapper{Type: l.Type},
			ExecAction:  *l.ExecAction,
		}, nil
	default:
		return nil, errors.New("unrecognized lifecycle handler type")
	}
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for LifecycleHandler.
func (l *LifecycleHandler) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var handlerType TypeWrapper
	err := unmarshal(&handlerType)
	if err != nil {
		return err
	}

	l.Type = handlerType.Type
	switch l.Type {
	case TypeHTTP:
		handler := &HTTPGetAction{}
		err = unmarshal(handler)
		l.HTTPGetAction = handler
	case TypeExec:
		handler := &ExecAction{}
		err = unmarshal(handler)
		l.ExecAction = handler
	default:
		return errors.New("unrecognized lifecycle handler type")
	}

	return err
}

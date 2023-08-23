package container

import (
	"encoding/json"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestContainerMarshalJSON(t *testing.T) {
	cases := []struct {
		input  Container
		result string
	}{
		{
			input: Container{
				Image: "nginx:v1",
			},
			result: `{"image":"nginx:v1"}`,
		},
		{
			input: Container{
				Image: "nginx:v1",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						ProbeType: ProbeType{Type: "Http"},
						HTTPGetAction: &HTTPGetAction{
							URL: "http://localhost:80",
						},
					},
					InitialDelaySeconds: 10,
				},
			},
			result: `{"image":"nginx:v1","readinessProbe":{"probeHandler":{"_type":"Http","url":"http://localhost:80"},"initialDelaySeconds":10}}`,
		},
		{
			input: Container{
				Image: "nginx:v1",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						ProbeType: ProbeType{Type: "Exec"},
						ExecAction: &ExecAction{
							Command: []string{"cat", "/tmp/healthy"},
						},
					},
					InitialDelaySeconds: 10,
				},
			},
			result: `{"image":"nginx:v1","readinessProbe":{"probeHandler":{"_type":"Exec","command":["cat","/tmp/healthy"]},"initialDelaySeconds":10}}`,
		},
		{
			input: Container{
				Image: "nginx:v1",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						ProbeType: ProbeType{Type: "Tcp"},
						TCPSocketAction: &TCPSocketAction{
							URL: "127.0.0.1:8080",
						},
					},
					InitialDelaySeconds: 10,
				},
			},
			result: `{"image":"nginx:v1","readinessProbe":{"probeHandler":{"_type":"Tcp","url":"127.0.0.1:8080"},"initialDelaySeconds":10}}`,
		},
	}

	for _, c := range cases {
		result, err := json.Marshal(&c.input)
		if err != nil {
			t.Errorf("Failed to marshal input: '%v': %v", c.input, err)
		}
		if string(result) != c.result {
			t.Errorf("Failed to marshal input: '%v': expected %+v, got %q", c.input, c.result, string(result))
		}
	}
}

func TestContainerUnmarshalJSON(t *testing.T) {
	cases := []struct {
		input  string
		result Container
	}{
		{
			input: `{"image":"nginx:v1"}`,
			result: Container{
				Image: "nginx:v1",
			},
		},
		{
			input: `{"image":"nginx:v1","readinessProbe":{"probeHandler":{"_type":"Http","url":"http://localhost:80"},"initialDelaySeconds":10}}`,
			result: Container{
				Image: "nginx:v1",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						ProbeType: ProbeType{Type: "Http"},
						HTTPGetAction: &HTTPGetAction{
							URL: "http://localhost:80",
						},
					},
					InitialDelaySeconds: 10,
				},
			},
		},
		{
			input: `{"image":"nginx:v1","readinessProbe":{"probeHandler":{"_type":"Exec","command":["cat","/tmp/healthy"]},"initialDelaySeconds":10}}`,
			result: Container{
				Image: "nginx:v1",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						ProbeType: ProbeType{Type: "Exec"},
						ExecAction: &ExecAction{
							Command: []string{"cat", "/tmp/healthy"},
						},
					},
					InitialDelaySeconds: 10,
				},
			},
		},
		{
			input: `{"image":"nginx:v1","readinessProbe":{"probeHandler":{"_type":"Tcp","url":"127.0.0.1:8080"},"initialDelaySeconds":10}}`,
			result: Container{
				Image: "nginx:v1",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						ProbeType: ProbeType{Type: "Tcp"},
						TCPSocketAction: &TCPSocketAction{
							URL: "127.0.0.1:8080",
						},
					},
					InitialDelaySeconds: 10,
				},
			},
		},
	}

	for _, c := range cases {
		var result Container
		if err := json.Unmarshal([]byte(c.input), &result); err != nil {
			t.Errorf("Failed to unmarshal input '%v': %v", c.input, err)
		}
		if !reflect.DeepEqual(result, c.result) {
			t.Errorf("Failed to unmarshal input '%v': expected %+v, got %+v", c.input, c.result, result)
		}
	}
}

func TestContainerMarshalYAML(t *testing.T) {
	cases := []struct {
		input  Container
		result string
	}{
		{
			input: Container{
				Image:   "nginx:v1",
				Command: []string{"/bin/sh", "-c", "echo hi"},
				Args:    []string{"/bin/sh", "-c", "echo hi"},
				Env: map[string]string{
					"env1": "VALUE",
				},
				WorkingDir: "/tmp",
			},
			result: `image: nginx:v1
command:
- /bin/sh
- -c
- echo hi
args:
- /bin/sh
- -c
- echo hi
env:
  env1: VALUE
workingDir: /tmp
`,
		},
		{
			input: Container{
				Image:   "nginx:v1",
				Command: []string{"/bin/sh", "-c", "echo hi"},
				Args:    []string{"/bin/sh", "-c", "echo hi"},
				Env: map[string]string{
					"env1": "VALUE",
				},
				WorkingDir: "/tmp",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						ProbeType: ProbeType{Type: "Http"},
						HTTPGetAction: &HTTPGetAction{
							URL: "http://localhost:80",
						},
					},
					InitialDelaySeconds: 10,
				},
			},
			result: `image: nginx:v1
command:
- /bin/sh
- -c
- echo hi
args:
- /bin/sh
- -c
- echo hi
env:
  env1: VALUE
workingDir: /tmp
readinessProbe:
  probeHandler:
    _type: Http
    url: http://localhost:80
  initialDelaySeconds: 10
`,
		},
		{
			input: Container{
				Image:   "nginx:v1",
				Command: []string{"/bin/sh", "-c", "echo hi"},
				Args:    []string{"/bin/sh", "-c", "echo hi"},
				Env: map[string]string{
					"env1": "VALUE",
				},
				WorkingDir: "/tmp",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						ProbeType: ProbeType{Type: "Exec"},
						ExecAction: &ExecAction{
							Command: []string{"cat", "/tmp/healthy"},
						},
					},
					InitialDelaySeconds: 10,
				},
			},
			result: `image: nginx:v1
command:
- /bin/sh
- -c
- echo hi
args:
- /bin/sh
- -c
- echo hi
env:
  env1: VALUE
workingDir: /tmp
readinessProbe:
  probeHandler:
    _type: Exec
    command:
    - cat
    - /tmp/healthy
  initialDelaySeconds: 10
`,
		},
		{
			input: Container{
				Image:   "nginx:v1",
				Command: []string{"/bin/sh", "-c", "echo hi"},
				Args:    []string{"/bin/sh", "-c", "echo hi"},
				Env: map[string]string{
					"env1": "VALUE",
				},
				WorkingDir: "/tmp",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						ProbeType: ProbeType{Type: "Tcp"},
						TCPSocketAction: &TCPSocketAction{
							URL: "127.0.0.1:8080",
						},
					},
					InitialDelaySeconds: 10,
				},
			},
			result: `image: nginx:v1
command:
- /bin/sh
- -c
- echo hi
args:
- /bin/sh
- -c
- echo hi
env:
  env1: VALUE
workingDir: /tmp
readinessProbe:
  probeHandler:
    _type: Tcp
    url: 127.0.0.1:8080
  initialDelaySeconds: 10
`,
		},
	}

	for _, c := range cases {
		result, err := yaml.Marshal(&c.input)
		if err != nil {
			t.Errorf("Failed to marshal input: '%v': %v", c.input, err)
		}
		if string(result) != c.result {
			t.Errorf("Failed to marshal input: '%v': expected %+v, got %q", c.input, c.result, string(result))
		}
	}
}

func TestContainerUnmarshalYAML(t *testing.T) {
	cases := []struct {
		input  string
		result Container
	}{
		{
			input: `image: nginx:v1
command:
- /bin/sh
- -c
- echo hi
args:
- /bin/sh
- -c
- echo hi
env:
  env1: VALUE
workingDir: /tmp
`,
			result: Container{
				Image:   "nginx:v1",
				Command: []string{"/bin/sh", "-c", "echo hi"},
				Args:    []string{"/bin/sh", "-c", "echo hi"},
				Env: map[string]string{
					"env1": "VALUE",
				},
				WorkingDir: "/tmp",
			},
		},
		{
			input: `image: nginx:v1
command:
- /bin/sh
- -c
- echo hi
args:
- /bin/sh
- -c
- echo hi
env:
  env1: VALUE
workingDir: /tmp
readinessProbe:
  probeHandler:
    _type: Http
    url: http://localhost:80
  initialDelaySeconds: 10
`,
			result: Container{
				Image:   "nginx:v1",
				Command: []string{"/bin/sh", "-c", "echo hi"},
				Args:    []string{"/bin/sh", "-c", "echo hi"},
				Env: map[string]string{
					"env1": "VALUE",
				},
				WorkingDir: "/tmp",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						ProbeType: ProbeType{Type: "Http"},
						HTTPGetAction: &HTTPGetAction{
							URL: "http://localhost:80",
						},
					},
					InitialDelaySeconds: 10,
				},
			},
		},
		{
			input: `image: nginx:v1
command:
- /bin/sh
- -c
- echo hi
args:
- /bin/sh
- -c
- echo hi
env:
  env1: VALUE
workingDir: /tmp
readinessProbe:
  probeHandler:
    _type: Exec
    command:
    - cat
    - /tmp/healthy
  initialDelaySeconds: 10
`,
			result: Container{
				Image:   "nginx:v1",
				Command: []string{"/bin/sh", "-c", "echo hi"},
				Args:    []string{"/bin/sh", "-c", "echo hi"},
				Env: map[string]string{
					"env1": "VALUE",
				},
				WorkingDir: "/tmp",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						ProbeType: ProbeType{Type: "Exec"},
						ExecAction: &ExecAction{
							Command: []string{"cat", "/tmp/healthy"},
						},
					},
					InitialDelaySeconds: 10,
				},
			},
		},
		{
			input: `image: nginx:v1
command:
- /bin/sh
- -c
- echo hi
args:
- /bin/sh
- -c
- echo hi
env:
  env1: VALUE
workingDir: /tmp
readinessProbe:
  probeHandler:
    _type: Tcp
    url: 127.0.0.1:8080
  initialDelaySeconds: 10
`,
			result: Container{
				Image:   "nginx:v1",
				Command: []string{"/bin/sh", "-c", "echo hi"},
				Args:    []string{"/bin/sh", "-c", "echo hi"},
				Env: map[string]string{
					"env1": "VALUE",
				},
				WorkingDir: "/tmp",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						ProbeType: ProbeType{Type: "Tcp"},
						TCPSocketAction: &TCPSocketAction{
							URL: "127.0.0.1:8080",
						},
					},
					InitialDelaySeconds: 10,
				},
			},
		},
	}

	for _, c := range cases {
		var result Container
		if err := yaml.Unmarshal([]byte(c.input), &result); err != nil {
			t.Errorf("Failed to unmarshal input '%v': %v", c.input, err)
		}
		if !reflect.DeepEqual(result, c.result) {
			t.Errorf("Failed to unmarshal input '%v': expected %+v, got %+v", c.input, c.result, result)
		}
	}
}

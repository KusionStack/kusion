package v1

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
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
				Resources: map[string]string{
					"cpu":    "4",
					"memory": "8Gi",
				},
				Files: map[string]FileSpec{
					"/tmp/test.txt": {
						Content: "hello world",
						Mode:    "0644",
					},
				},
			},
			result: `{"image":"nginx:v1","resources":{"cpu":"4","memory":"8Gi"},"files":{"/tmp/test.txt":{"content":"hello world","mode":"0644"}}}`,
		},
		{
			input: Container{
				Image: "nginx:v1",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						TypeWrapper: TypeWrapper{"Http"},
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
						TypeWrapper: TypeWrapper{"Exec"},
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
						TypeWrapper: TypeWrapper{Type: "Tcp"},
						TCPSocketAction: &TCPSocketAction{
							URL: "127.0.0.1:8080",
						},
					},
					InitialDelaySeconds: 10,
				},
			},
			result: `{"image":"nginx:v1","readinessProbe":{"probeHandler":{"_type":"Tcp","url":"127.0.0.1:8080"},"initialDelaySeconds":10}}`,
		},
		{
			input: Container{
				Image: "nginx:v1",
				Lifecycle: &Lifecycle{
					PostStart: &LifecycleHandler{
						TypeWrapper: TypeWrapper{"Exec"},
						ExecAction: &ExecAction{
							Command: []string{"/bin/sh", "-c", "nginx -s quit; while killall -0 nginx; do sleep 1; done"},
						},
					},
					PreStop: &LifecycleHandler{
						TypeWrapper: TypeWrapper{"Exec"},
						ExecAction: &ExecAction{
							Command: []string{"/bin/sh", "-c", "echo Hello from the postStart handler > /usr/share/message"},
						},
					},
				},
			},
			result: `{"image":"nginx:v1","lifecycle":{"preStop":{"_type":"Exec","command":["/bin/sh","-c","echo Hello from the postStart handler \u003e /usr/share/message"]},"postStart":{"_type":"Exec","command":["/bin/sh","-c","nginx -s quit; while killall -0 nginx; do sleep 1; done"]}}}`,
		},
		{
			input: Container{
				Image: "nginx:v1",
				Lifecycle: &Lifecycle{
					PostStart: &LifecycleHandler{
						TypeWrapper: TypeWrapper{"Http"},
						HTTPGetAction: &HTTPGetAction{
							URL: "http://localhost:80",
						},
					},
					PreStop: &LifecycleHandler{
						TypeWrapper: TypeWrapper{"Http"},
						HTTPGetAction: &HTTPGetAction{
							URL: "http://localhost:80",
						},
					},
				},
			},
			result: `{"image":"nginx:v1","lifecycle":{"preStop":{"_type":"Http","url":"http://localhost:80"},"postStart":{"_type":"Http","url":"http://localhost:80"}}}`,
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
				Env: yaml.MapSlice{
					{
						Key:   "env1",
						Value: "VALUE",
					},
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
				Env: yaml.MapSlice{
					{
						Key:   "env1",
						Value: "VALUE",
					},
				},
				WorkingDir: "/tmp",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						TypeWrapper: TypeWrapper{Type: "Http"},
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
				Env: yaml.MapSlice{
					{
						Key:   "env1",
						Value: "VALUE",
					},
				},
				WorkingDir: "/tmp",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						TypeWrapper: TypeWrapper{Type: "Exec"},
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
				Env: yaml.MapSlice{
					{
						Key:   "env1",
						Value: "VALUE",
					},
				},
				WorkingDir: "/tmp",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						TypeWrapper: TypeWrapper{Type: "Tcp"},
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
		{
			input: Container{
				Image:   "nginx:v1",
				Command: []string{"/bin/sh", "-c", "echo hi"},
				Args:    []string{"/bin/sh", "-c", "echo hi"},
				Env: yaml.MapSlice{
					{
						Key:   "env1",
						Value: "VALUE",
					},
				},
				WorkingDir: "/tmp",
				Lifecycle: &Lifecycle{
					PostStart: &LifecycleHandler{
						TypeWrapper: TypeWrapper{"Exec"},
						ExecAction: &ExecAction{
							Command: []string{"/bin/sh", "-c", "nginx -s quit; while killall -0 nginx; do sleep 1; done"},
						},
					},
					PreStop: &LifecycleHandler{
						TypeWrapper: TypeWrapper{"Exec"},
						ExecAction: &ExecAction{
							Command: []string{"/bin/sh", "-c", "echo Hello from the postStart handler > /usr/share/message"},
						},
					},
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
lifecycle:
  preStop:
    _type: Exec
    command:
    - /bin/sh
    - -c
    - echo Hello from the postStart handler > /usr/share/message
  postStart:
    _type: Exec
    command:
    - /bin/sh
    - -c
    - nginx -s quit; while killall -0 nginx; do sleep 1; done
`,
		},
		{
			input: Container{
				Image:   "nginx:v1",
				Command: []string{"/bin/sh", "-c", "echo hi"},
				Args:    []string{"/bin/sh", "-c", "echo hi"},
				Env: yaml.MapSlice{
					{
						Key:   "env1",
						Value: "VALUE",
					},
				},
				WorkingDir: "/tmp",
				Lifecycle: &Lifecycle{
					PostStart: &LifecycleHandler{
						TypeWrapper: TypeWrapper{"Http"},
						HTTPGetAction: &HTTPGetAction{
							URL: "http://localhost:80",
						},
					},
					PreStop: &LifecycleHandler{
						TypeWrapper: TypeWrapper{"Http"},
						HTTPGetAction: &HTTPGetAction{
							URL: "http://localhost:80",
						},
					},
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
lifecycle:
  preStop:
    _type: Http
    url: http://localhost:80
  postStart:
    _type: Http
    url: http://localhost:80
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

func TestWorkloadMarshalJSON(t *testing.T) {
	r2 := int32(2)
	tests := []struct {
		name          string
		data          *Workload
		expected      string
		expectedError error
	}{
		{
			name: "Valid MarshalJSON for Service",
			data: &Workload{
				Header: Header{
					Type: TypeService,
				},
				Service: &Service{
					Type: "Deployment",
					Base: Base{
						Replicas: &r2,
						Labels: map[string]string{
							"app": "my-service",
						},
					},
				},
				Job: &Job{
					Schedule: "* * * * *",
				},
			},
			expected:      `{"_type": "kam.v1.workload.Service", "replicas": 2, "labels": {"app": "my-service"}, "type": "Deployment"}`,
			expectedError: nil,
		},
		{
			name: "Valid MarshalJSON for Job",
			data: &Workload{
				Header: Header{
					Type: TypeJob,
				},
				Job: &Job{
					Schedule: "* * * * *",
				},
			},
			expected:      `{"_type": "kam.v1.workload.Job", "schedule": "* * * * *"}`,
			expectedError: nil,
		},
		{
			name: "Unknown _Type",
			data: &Workload{
				Header: Header{
					Type: "Unknown",
				},
				Job: &Job{
					Schedule: "* * * * *",
				},
			},
			expected:      "",
			expectedError: errors.New("unknown workload type"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, actualErr := json.Marshal(test.data)
			if test.expectedError == nil {
				assert.JSONEq(t, test.expected, string(actual))
				assert.NoError(t, actualErr)
			} else {
				assert.ErrorContains(t, actualErr, test.expectedError.Error())
			}
		})
	}
}

func TestWorkloadMarshalYAML(t *testing.T) {
	r2 := int32(2)
	tests := []struct {
		name          string
		workload      *Workload
		expected      string
		expectedError error
	}{
		{
			name: "Valid MarshalYAML for Service",
			workload: &Workload{
				Header: Header{
					Type: TypeService,
				},
				Service: &Service{
					Type: "Deployment",
					Base: Base{
						Replicas: &r2,
						Labels: map[string]string{
							"app": "my-service",
						},
					},
				},
			},
			expected:      `{"_type":"kam.v1.workload.Service", "replicas": 2, "labels": {"app": "my-service"}, "type": "Deployment"}`,
			expectedError: nil,
		},
		{
			name: "Valid MarshalYAML for Job",
			workload: &Workload{
				Header: Header{
					Type: TypeJob,
				},
				Job: &Job{
					Schedule: "* * * * *",
				},
			},
			expected:      `{"_type":"kam.v1.workload.Job", "schedule": "* * * * *"}`,
			expectedError: nil,
		},
		{
			name: "Unknown _Type",
			workload: &Workload{
				Header: Header{
					Type: "Unknown",
				},
				Job: &Job{
					Schedule: "* * * * *",
				},
			},
			expected:      "",
			expectedError: errors.New("unknown workload type"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, actualErr := yaml.Marshal(test.workload)
			if test.expectedError == nil {
				assert.YAMLEq(t, test.expected, string(actual))
				assert.NoError(t, actualErr)
			} else {
				assert.ErrorContains(t, actualErr, test.expectedError.Error())
			}
		})
	}
}

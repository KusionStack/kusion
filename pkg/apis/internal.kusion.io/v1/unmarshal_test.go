package v1

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestContainerUnmarshalJSON(t *testing.T) {
	cases := []struct {
		input  string
		result Container
	}{
		{
			input: `{"image":"nginx:v1","resources":{"cpu":"4","memory":"8Gi"},"files":{"/tmp/test.txt":{"content":"hello world","mode":"0644"}}}`,
			result: Container{
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
		},
		{
			input: `{"image":"nginx:v1","readinessProbe":{"probeHandler":{"_type":"service.container.probe.Http","url":"http://localhost:80"},"initialDelaySeconds":10}}`,
			result: Container{
				Image: "nginx:v1",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						TypeWrapper: TypeWrapper{Type: "service.container.probe.Http"},
						HTTPGetAction: &HTTPGetAction{
							URL: "http://localhost:80",
						},
					},
					InitialDelaySeconds: 10,
				},
			},
		},
		{
			input: `{"image":"nginx:v1","readinessProbe":{"probeHandler":{"_type":"service.container.probe.Exec","command":["cat","/tmp/healthy"]},"initialDelaySeconds":10}}`,
			result: Container{
				Image: "nginx:v1",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						TypeWrapper: TypeWrapper{Type: "service.container.probe.Exec"},
						ExecAction: &ExecAction{
							Command: []string{"cat", "/tmp/healthy"},
						},
					},
					InitialDelaySeconds: 10,
				},
			},
		},
		{
			input: `{"image":"nginx:v1","readinessProbe":{"probeHandler":{"_type":"service.container.probe.Tcp","url":"127.0.0.1:8080"},"initialDelaySeconds":10}}`,
			result: Container{
				Image: "nginx:v1",
				ReadinessProbe: &Probe{
					ProbeHandler: &ProbeHandler{
						TypeWrapper: TypeWrapper{Type: "service.container.probe.Tcp"},
						TCPSocketAction: &TCPSocketAction{
							URL: "127.0.0.1:8080",
						},
					},
					InitialDelaySeconds: 10,
				},
			},
		},
		{
			input: `{"image":"nginx:v1","lifecycle":{"preStop":{"_type":"service.container.probe.Exec","command":["/bin/sh","-c","echo Hello from the postStart handler \u003e /usr/share/message"]},"postStart":{"_type":"service.container.probe.Exec","command":["/bin/sh","-c","nginx -s quit; while killall -0 nginx; do sleep 1; done"]}}}`,
			result: Container{
				Image: "nginx:v1",
				Lifecycle: &Lifecycle{
					PostStart: &LifecycleHandler{
						TypeWrapper: TypeWrapper{"service.container.probe.Exec"},
						ExecAction: &ExecAction{
							Command: []string{"/bin/sh", "-c", "nginx -s quit; while killall -0 nginx; do sleep 1; done"},
						},
					},
					PreStop: &LifecycleHandler{
						TypeWrapper: TypeWrapper{"service.container.probe.Exec"},
						ExecAction: &ExecAction{
							Command: []string{"/bin/sh", "-c", "echo Hello from the postStart handler > /usr/share/message"},
						},
					},
				},
			},
		},
		{
			input: `{"image":"nginx:v1","lifecycle":{"preStop":{"_type":"service.container.probe.Http","url":"http://localhost:80"},"postStart":{"_type":"service.container.probe.Http","url":"http://localhost:80"}}}`,
			result: Container{
				Image: "nginx:v1",
				Lifecycle: &Lifecycle{
					PostStart: &LifecycleHandler{
						TypeWrapper: TypeWrapper{"service.container.probe.Http"},
						HTTPGetAction: &HTTPGetAction{
							URL: "http://localhost:80",
						},
					},
					PreStop: &LifecycleHandler{
						TypeWrapper: TypeWrapper{"service.container.probe.Http"},
						HTTPGetAction: &HTTPGetAction{
							URL: "http://localhost:80",
						},
					},
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
				Env: yaml.MapSlice{
					{
						Key:   "env1",
						Value: "VALUE",
					},
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
    _type: service.container.probe.Http
    url: http://localhost:80
  initialDelaySeconds: 10
`,
			result: Container{
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
						TypeWrapper: TypeWrapper{Type: "service.container.probe.Http"},
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
    _type: service.container.probe.Exec
    command:
    - cat
    - /tmp/healthy
  initialDelaySeconds: 10
`,
			result: Container{
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
						TypeWrapper: TypeWrapper{Type: "service.container.probe.Exec"},
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
    _type: service.container.probe.Tcp
    url: 127.0.0.1:8080
  initialDelaySeconds: 10
`,
			result: Container{
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
						TypeWrapper: TypeWrapper{Type: "service.container.probe.Tcp"},
						TCPSocketAction: &TCPSocketAction{
							URL: "127.0.0.1:8080",
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
lifecycle:
  preStop:
    _type: service.container.probe.Exec
    command:
    - /bin/sh
    - -c
    - echo Hello from the postStart handler > /usr/share/message
  postStart:
    _type: service.container.probe.Exec
    command:
    - /bin/sh
    - -c
    - nginx -s quit; while killall -0 nginx; do sleep 1; done
`,
			result: Container{
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
						TypeWrapper: TypeWrapper{"service.container.probe.Exec"},
						ExecAction: &ExecAction{
							Command: []string{"/bin/sh", "-c", "nginx -s quit; while killall -0 nginx; do sleep 1; done"},
						},
					},
					PreStop: &LifecycleHandler{
						TypeWrapper: TypeWrapper{"service.container.probe.Exec"},
						ExecAction: &ExecAction{
							Command: []string{"/bin/sh", "-c", "echo Hello from the postStart handler > /usr/share/message"},
						},
					},
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
lifecycle:
  preStop:
    _type: service.container.probe.Http
    url: http://localhost:80
  postStart:
    _type: service.container.probe.Http
    url: http://localhost:80
`,
			result: Container{
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
						TypeWrapper: TypeWrapper{"service.container.probe.Http"},
						HTTPGetAction: &HTTPGetAction{
							URL: "http://localhost:80",
						},
					},
					PreStop: &LifecycleHandler{
						TypeWrapper: TypeWrapper{"service.container.probe.Http"},
						HTTPGetAction: &HTTPGetAction{
							URL: "http://localhost:80",
						},
					},
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

func TestWorkloadUnmarshalJSON(t *testing.T) {
	r1 := int32(1)
	tests := []struct {
		name          string
		data          string
		expected      Workload
		expectedError error
	}{
		{
			name: "Valid UnmarshalJSON for Service",
			data: `{"_type": "service.Service", "replicas": 1, "labels": {}, "annotations": {}, "dirs": {}, "schedule": "* * * * *"}`,
			expected: Workload{
				Header: Header{
					Type: TypeService,
				},
				Service: &Service{
					Base: Base{
						Replicas:    &r1,
						Labels:      map[string]string{},
						Annotations: map[string]string{},
						Dirs:        map[string]string{},
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "Valid UnmarshalJSON for Job",
			data: `{"_type": "job.Job", "schedule": "* * * * *"}`,
			expected: Workload{
				Header: Header{
					Type: TypeJob,
				},
				Job: &Job{
					Schedule: "* * * * *",
				},
			},
			expectedError: nil,
		},
		{
			name: "Unknown _Type",
			data: `{"_type": "Unknown", "replicas": 1, "labels": {}, "annotations": {}, "dirs": {}, "schedule": "* * * * *"}`,
			expected: Workload{
				Header: Header{
					Type: "Unknown",
				},
			},
			expectedError: errors.New("unknown workload type"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual Workload
			actualErr := json.Unmarshal([]byte(test.data), &actual)
			if test.expectedError == nil {
				assert.Equal(t, test.expected, actual)
				assert.NoError(t, actualErr)
			} else {
				assert.ErrorContains(t, actualErr, test.expectedError.Error())
			}
		})
	}
}

func TestWorkloadUnmarshalYAML(t *testing.T) {
	r1 := int32(1)
	tests := []struct {
		name          string
		data          string
		expected      Workload
		expectedError error
	}{
		{
			name: "Valid UnmarshalYAML for Service",
			data: `_type: service.Service
replicas: 1
labels: {}
annotations: {}
dirs: {}
schedule: '* * * * *'`,
			expected: Workload{
				Header: Header{
					Type: TypeService,
				},
				Service: &Service{
					Base: Base{
						Replicas:    &r1,
						Labels:      map[string]string{},
						Annotations: map[string]string{},
						Dirs:        map[string]string{},
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "Valid UnmarshalYAML for Job",
			data: `_type: job.Job
replicas: 1
labels: {}
annotations: {}
dirs: {}
schedule: '* * * * *'`,
			expected: Workload{
				Header: Header{
					Type: TypeJob,
				},
				Job: &Job{
					Base: Base{
						Replicas:    &r1,
						Labels:      map[string]string{},
						Annotations: map[string]string{},
						Dirs:        map[string]string{},
					},
					Schedule: "* * * * *",
				},
			},
			expectedError: nil,
		},
		{
			name: "Unknown _Type",
			data: `_type: Unknown
replicas: 1
labels: {}
annotations: {}
dirs: {}
schedule: '* * * * *'`,
			expected:      Workload{},
			expectedError: errors.New("unknown workload type"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual Workload
			actualErr := yaml.Unmarshal([]byte(test.data), &actual)
			if test.expectedError == nil {
				assert.Equal(t, test.expected, actual)
				assert.NoError(t, actualErr)
			} else {
				assert.ErrorContains(t, actualErr, test.expectedError.Error())
			}
		})
	}
}

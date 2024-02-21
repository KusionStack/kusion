package workload

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestWorkload_MarshalJSON(t *testing.T) {
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
						Replicas: 2,
						Labels: map[string]string{
							"app": "my-service",
						},
					},
				},
				Job: &Job{
					Schedule: "* * * * *",
				},
			},
			expected:      `{"_type": "Service", "replicas": 2, "labels": {"app": "my-service"}, "type": "Deployment"}`,
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
			expected:      `{"_type": "Job", "schedule": "* * * * *"}`,
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

func TestWorkload_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name          string
		data          string
		expected      Workload
		expectedError error
	}{
		{
			name: "Valid UnmarshalJSON for Service",
			data: `{"_type": "Service", "replicas": 1, "labels": {}, "annotations": {}, "dirs": {}, "schedule": "* * * * *"}`,
			expected: Workload{
				Header: Header{
					Type: TypeService,
				},
				Service: &Service{
					Base: Base{
						Replicas:    1,
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
			data: `{"_type": "Job", "schedule": "* * * * *"}`,
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

func TestWorkload_MarshalYAML(t *testing.T) {
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
						Replicas: 2,
						Labels: map[string]string{
							"app": "my-service",
						},
					},
				},
				Job: &Job{
					Schedule: "* * * * *",
				},
			},
			expected: `_type: Service
replicas: 2
labels:
    app: my-service
type: Deployment`,
			expectedError: nil,
		},
		{
			name: "Valid MarshalYAML for Job",
			workload: &Workload{
				Header: Header{
					Type: TypeJob,
				},
				Service: &Service{
					Type: "Deployment",
					Base: Base{
						Replicas: 2,
						Labels: map[string]string{
							"app": "my-service",
						},
					},
				},
				Job: &Job{
					Schedule: "* * * * *",
				},
			},
			expected: `_type: Job
schedule: '* * * * *'`,
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

func TestWorkload_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name          string
		data          string
		expected      Workload
		expectedError error
	}{
		{
			name: "Valid UnmarshalYAML for Service",
			data: `_type: Service
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
						Replicas:    1,
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
			data: `_type: Job
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
						Replicas:    1,
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

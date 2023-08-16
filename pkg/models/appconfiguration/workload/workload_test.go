package workload

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestWorkload_MarshalJSON(t *testing.T) {
	data := Workload{
		WorkloadHeader: WorkloadHeader{
			Type: WorkloadTypeService,
		},
		Service: &Service{
			WorkloadBase: WorkloadBase{
				Replicas: 2,
				Labels: map[string]string{
					"app": "my-service",
				},
			},
		},
		Job: &Job{
			Schedule: "* * * * *",
		},
	}

	expected := `{"_type":"Service","replicas":2,"labels":{"app":"my-service"}}`
	actual, err := json.Marshal(data)
	if err != nil {
		t.Errorf("Error while marshaling workload: %v", err)
	}

	if string(actual) != expected {
		t.Errorf("Expected marshaled JSON: %s, got: %s", expected, string(actual))
	}
}

func TestWorkload_UnmarshalJSON(t *testing.T) {
	data := `{"_type":"Service","replicas":1,"labels":{},"annotations":{},"dirs":{},"schedule":"* * * * *"}`

	expected := Workload{
		WorkloadHeader: WorkloadHeader{
			Type: WorkloadTypeService,
		},
		Service: &Service{
			WorkloadBase: WorkloadBase{
				Replicas:    1,
				Labels:      map[string]string{},
				Annotations: map[string]string{},
				Dirs:        map[string]string{},
			},
		},
	}
	var actual Workload
	err := json.Unmarshal([]byte(data), &actual)
	if err != nil {
		t.Errorf("Error while unmarshaling JSON: %v", err)
	}

	if actual.Type != expected.Type {
		t.Errorf("Expected workload type: %s, got: %s", expected.Type, actual.Type)
	}

	if actual.Service == nil {
		t.Errorf("Expected service is not nil, got: %v", expected.Service)
	}

	if actual.Job != nil {
		t.Errorf("Expected job is nil, got: %v", expected.Job)
	}
}

func TestWorkload_UnmarshalJSON_UnknownType(t *testing.T) {
	data := `{"_type":"Unknown","replicas":1,"labels":{},"annotations":{},"dirs":{},"schedule":"* * * * *"}`

	var actual Workload
	actualErr := json.Unmarshal([]byte(data), &actual)
	if actualErr == nil {
		t.Error("Expected error for unknown workload type")
	}

	expectedError := errors.New("unknown workload type")
	if actualErr.Error() != expectedError.Error() {
		t.Errorf("Expected error: %v, got: %v", expectedError, actualErr)
	}
}

func TestWorkload_MarshalYAML(t *testing.T) {
	data := Workload{
		WorkloadHeader: WorkloadHeader{
			Type: WorkloadTypeService,
		},
		Service: &Service{
			WorkloadBase: WorkloadBase{
				Replicas: 2,
				Labels: map[string]string{
					"app": "my-service",
				},
			},
		},
		Job: &Job{
			Schedule: "* * * * *",
		},
	}

	expected := `_type: Service
replicas: 2
labels:
    app: my-service
`
	actual, err := yaml.Marshal(data)
	if err != nil {
		t.Errorf("Error while marshaling workload: %v", err)
	}
	if string(actual) != expected {
		t.Errorf("Expected marshaled YAML:\n%s\ngot:\n%s", expected, string(actual))
	}
}

func TestWorkload_UnmarshalYAML(t *testing.T) {
	data := `_type: Service
replicas: 1
labels: {}
annotations: {}
dirs: {}
schedule: "* * * * *"
`

	expected := Workload{
		WorkloadHeader: WorkloadHeader{
			Type: WorkloadTypeService,
		},
		Service: &Service{
			WorkloadBase: WorkloadBase{
				Replicas:    1,
				Labels:      map[string]string{},
				Annotations: map[string]string{},
				Dirs:        map[string]string{},
			},
		},
	}
	var actual Workload
	err := yaml.Unmarshal([]byte(data), &actual)
	if err != nil {
		t.Errorf("Error while unmarshaling YAML: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Unexpected YAML deserialization result.\nExpected: %v\nActual: %v", expected, actual)
	}
}

func TestWorkload_UnmarshalYAML_UnknownType(t *testing.T) {
	data := `_type: Unknown
replicas: 1
labels: {}
annotations: {}
dirs: {}
schedule: "* * * * *"
`

	var actual Workload
	actualErr := yaml.Unmarshal([]byte(data), &actual)
	expectedError := errors.New("unknown workload type")
	if actualErr == nil || actualErr.Error() != expectedError.Error() {
		t.Errorf("Expected error: %v, got: %v", expectedError, actualErr)
	}
}

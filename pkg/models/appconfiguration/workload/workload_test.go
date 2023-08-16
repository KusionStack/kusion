package workload

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestWorkload_MarshalJSON(t *testing.T) {
	workload := Workload{
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

	data, err := json.Marshal(workload)
	if err != nil {
		t.Errorf("Error while marshaling workload: %v", err)
	}

	if string(data) != expected {
		t.Errorf("Expected marshaled JSON: %s, got: %s", expected, string(data))
	}
}

func TestWorkload_UnmarshalJSON(t *testing.T) {
	jsonData := `{"_type":"Service","replicas":1,"labels":{},"annotations":{},"dirs":{},"schedule":"* * * * *"}`
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
	err := json.Unmarshal([]byte(jsonData), &actual)
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
	jsonData := `{"_type":"Unknown","replicas":1,"labels":{},"annotations":{},"dirs":{},"schedule":"* * * * *"}`

	var workload Workload
	err := json.Unmarshal([]byte(jsonData), &workload)
	if err == nil {
		t.Error("Expected error for unknown workload type")
	}

	expectedError := errors.New("unknown workload type")
	if err.Error() != expectedError.Error() {
		t.Errorf("Expected error: %v, got: %v", expectedError, err)
	}
}

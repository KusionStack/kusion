package workload

import (
	"encoding/json"
	"errors"
)

type WorkloadType string

const (
	WorkloadTypeJob     = "Job"
	WorkloadTypeService = "Service"
)

type WorkloadHeader struct {
	Type WorkloadType `yaml:"_type" json:"_type"`
}

type Workload struct {
	WorkloadHeader `yaml:",inline" json:",inline"`
	*Service       `yaml:",inline" json:",inline"`
	*Job           `yaml:",inline" json:",inline"`
}

func (w Workload) MarshalJSON() ([]byte, error) {
	switch w.Type {
	case WorkloadTypeService:
		return json.Marshal(struct {
			WorkloadHeader `yaml:",inline" json:",inline"`
			*Service       `json:",inline"`
		}{
			WorkloadHeader: WorkloadHeader{w.Type},
			Service:        w.Service,
		})
	case WorkloadTypeJob:
		return json.Marshal(struct {
			WorkloadHeader `yaml:",inline" json:",inline"`
			*Job           `json:",inline"`
		}{
			WorkloadHeader: WorkloadHeader{w.Type},
			Job:            w.Job,
		})
	default:
		return nil, errors.New("unknown workload type")
	}
}

func (w *Workload) UnmarshalJSON(data []byte) error {
	var workloadData WorkloadHeader
	err := json.Unmarshal(data, &workloadData)
	if err != nil {
		return err
	}

	w.Type = workloadData.Type
	switch w.Type {
	case WorkloadTypeJob:
		var v Job
		err = json.Unmarshal(data, &v)
		w.Job = &v
	case WorkloadTypeService:
		var v Service
		err = json.Unmarshal(data, &v)
		w.Service = &v
	default:
		err = errors.New("unknown workload type")
	}

	return err
}

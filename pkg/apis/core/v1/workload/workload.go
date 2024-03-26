package workload

import (
	"encoding/json"
	"fmt"

	"kusionstack.io/kusion/pkg/apis/core"
)

type Type string

const (
	TypeJob       = core.BuiltinModulePrefix + "v1.workload.Job"
	TypeService   = core.BuiltinModulePrefix + "v1.workload.Service"
	FieldReplicas = "replicas"
)

type Header struct {
	Type string `yaml:"_type" json:"_type"`
}

type Workload struct {
	Header   `yaml:",inline" json:",inline"`
	*Service `yaml:",inline" json:",inline"`
	*Job     `yaml:",inline" json:",inline"`
}

func (w *Workload) MarshalJSON() ([]byte, error) {
	switch w.Header.Type {
	case TypeService:
		return json.Marshal(struct {
			Header   `yaml:",inline" json:",inline"`
			*Service `json:",inline"`
		}{
			Header:  Header{w.Header.Type},
			Service: w.Service,
		})
	case TypeJob:
		return json.Marshal(struct {
			Header `yaml:",inline" json:",inline"`
			*Job   `json:",inline"`
		}{
			Header: Header{w.Header.Type},
			Job:    w.Job,
		})
	default:
		return nil, fmt.Errorf("unknown workload type: %s", w.Header.Type)
	}
}

func (w *Workload) UnmarshalJSON(data []byte) error {
	var workloadData Header
	err := json.Unmarshal(data, &workloadData)
	if err != nil {
		return err
	}

	w.Header.Type = workloadData.Type
	switch w.Header.Type {
	case TypeJob:
		var v Job
		err = json.Unmarshal(data, &v)
		w.Job = &v
	case TypeService:
		var v Service
		err = json.Unmarshal(data, &v)
		w.Service = &v
	default:
		err = fmt.Errorf("unknown workload type: %s", w.Header.Type)
	}

	return err
}

func (w *Workload) MarshalYAML() (interface{}, error) {
	switch w.Header.Type {
	case TypeService:
		return struct {
			Header  `yaml:",inline" json:",inline"`
			Service `yaml:",inline" json:",inline"`
		}{
			Header:  Header{w.Header.Type},
			Service: *w.Service,
		}, nil
	case TypeJob:
		return struct {
			Header `yaml:",inline" json:",inline"`
			*Job   `yaml:",inline" json:",inline"`
		}{
			Header: Header{w.Header.Type},
			Job:    w.Job,
		}, nil
	default:
		return nil, fmt.Errorf("unknown workload type: %s", w.Header.Type)
	}
}

func (w *Workload) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var workloadData Header
	err := unmarshal(&workloadData)
	if err != nil {
		return err
	}

	w.Header.Type = workloadData.Type
	switch w.Header.Type {
	case TypeJob:
		var v Job
		err = unmarshal(&v)
		w.Job = &v
	case TypeService:
		var v Service
		err = unmarshal(&v)
		w.Service = &v
	default:
		err = fmt.Errorf("unknown workload type: %s", w.Header.Type)
	}

	return err
}

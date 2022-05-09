package dyff

import (
	"encoding/json"
	"reflect"
)

var (
	CustomComparatorMap = map[string]Comparator{
		"/spec/template/metadata/annotations/pod.beta1.sigma.ali/alloc-spec":           JsonStrComparator,
		"/spec/template/metadata/annotations/pod.beta1.sigma.ali/container-state-spec": JsonStrComparator,
	}
)

type Comparator func(from, to string) bool

func JsonStrComparator(from, to string) bool {
	fromJson := make(map[string]interface{})
	err := json.Unmarshal([]byte(from), &fromJson)
	if err != nil {
		return false
	}
	toJson := make(map[string]interface{})
	err = json.Unmarshal([]byte(to), &toJson)
	if err != nil {
		return false
	}
	return reflect.DeepEqual(fromJson, toJson)
}

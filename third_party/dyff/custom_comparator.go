package dyff

import (
	"encoding/json"
	"reflect"
)

var CustomComparatorMap = map[string]Comparator{
	"/spec/template/metadata/annotations/pod.beta1.sigma.ali/alloc-spec":           JSONStrComparator,
	"/spec/template/metadata/annotations/pod.beta1.sigma.ali/container-state-spec": JSONStrComparator,
}

type Comparator func(from, to string) bool

func JSONStrComparator(from, to string) bool {
	fromJSON := make(map[string]interface{})
	err := json.Unmarshal([]byte(from), &fromJSON)
	if err != nil {
		return false
	}
	toJSON := make(map[string]interface{})
	err = json.Unmarshal([]byte(to), &toJSON)
	if err != nil {
		return false
	}
	return reflect.DeepEqual(fromJSON, toJSON)
}

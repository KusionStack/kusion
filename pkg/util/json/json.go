package json

import (
	"encoding/json"

	"kusionstack.io/kusion/pkg/util"
)

// https://github.com/ksonnet/ksonnet/blob/master/pkg/kubecfg/diff.go
func removeFields(config, live interface{}) interface{} {
	switch c := config.(type) {
	case map[string]interface{}:
		l, ok := live.(map[string]interface{})
		if ok {
			return RemoveMapFields(c, l)
		} else {
			return live
		}
	case []interface{}:
		l, ok := live.([]interface{})
		if ok {
			return RemoveListFields(c, l)
		} else {
			return live
		}
	default:
		return live
	}
}

// RemoveMapFields remove all non-existent fields in the live that don't exist in the config
func RemoveMapFields(config, live map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for k, v1 := range config {
		v2, ok := live[k]
		if !ok {
			continue
		}
		if v2 != nil {
			v2 = removeFields(v1, v2)
		}
		result[k] = v2
	}
	return result
}

// If live is longer than config, then the extra elements at the end of the
// list will be returned as-is so they appear in the diff.
func RemoveListFields(config, live []interface{}) []interface{} {
	result := make([]interface{}, 0, len(live))
	for i, v2 := range live {
		if len(config) > i {
			if v2 != nil {
				v2 = removeFields(config[i], v2)
			}
			result = append(result, v2)
		} else {
			result = append(result, v2)
		}
	}
	return result
}

// Marshal2String marshal to string
func Marshal2String(v interface{}) string {
	r, err := json.Marshal(v)
	if err != nil {
		return "marshal failed"
	}

	return string(r)
}

// Marshal2PrettyString marshal to pretty string
func Marshal2PrettyString(v interface{}) string {
	r, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "marshal to pretty json failed"
	}

	return string(r)
}

// MustMarshal2String marshal to string and panic on error
func MustMarshal2String(v interface{}) string {
	r, err := json.Marshal(v)
	util.CheckNotError(err, "json marshal failed")

	return string(r)
}

// MustMarshal2PrettyString marshal to pretty string and panic on error
func MustMarshal2PrettyString(v interface{}) string {
	r, err := json.MarshalIndent(v, "", "  ")
	util.CheckNotError(err, "must marshal to pretty json failed")

	return string(r)
}

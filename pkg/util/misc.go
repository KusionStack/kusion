package util

import "reflect"

// InArray returns true if the string 'x' is found in the array 'array', else false
func InArray(x string, array []string) bool {
	if len(array) == 0 {
		return false
	}
	for _, y := range array {
		if x == y {
			return true
		}
	}
	return false
}

func IsZero(x interface{}) bool {
	if x == nil {
		return true
	}
	value := reflect.ValueOf(x)
	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	}
	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
}

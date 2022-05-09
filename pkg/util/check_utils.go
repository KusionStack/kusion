package util

import (
	"fmt"
	"reflect"
)

func CheckArgument(exp bool, msg string) {
	if !exp {
		panic(msg)
	}
}

func CheckNotNil(in interface{}, msg string) {
	if in == nil || (reflect.TypeOf(in).Kind() == reflect.Ptr && reflect.ValueOf(in).IsNil()) {
		panic(fmt.Sprintf("checkArgument failed: %v", msg))
	}
}

func CheckNotError(e error, msg string) {
	if e != nil {
		panic(fmt.Sprintf("error:%v, msg: %v", e.Error(), msg))
	}
}

package util

import (
	"errors"
)

func RecoverErr(err *error) {
	if r := recover(); r != nil {
		switch x := r.(type) {
		case string:
			*err = errors.New(x)
		case error:
			*err = x
		default:
			*err = errors.New("Unknow panic")
		}
	}
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

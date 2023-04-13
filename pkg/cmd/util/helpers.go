package util

import (
	"errors"
	"strings"
)

func RecoverErr(err *error) {
	if r := recover(); r != nil {
		switch x := r.(type) {
		case string:
			*err = errors.New(x)
		case error:
			*err = x
		default:
			*err = errors.New("unknown panic")
		}
	}
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

func ParseClusterArgument(args []string) string {
	var cluster string
	for _, argument := range args {
		// e.g. cluster=xxx
		if strings.HasPrefix(argument, "cluster=") {
			split := strings.Split(argument, "=")
			cluster = split[1]
		}
	}
	return cluster
}

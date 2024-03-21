package util

import (
	"github.com/pkg/errors"
)

func AggregateError(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	var errMsg string
	for _, err := range errs {
		if err != nil && err.Error() != "" {
			errMsg = errMsg + err.Error() + "; "
		}
	}
	if errMsg != "" {
		errMsg = errMsg[:len(errMsg)-2]
	}
	return errors.New(errMsg)
}

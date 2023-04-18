package svc

import (
	"kusionstack.io/kusion/pkg/status"
)

type Error struct {
	Code    status.Code `json:"code" yaml:"code"`
	Message string      `json:"message" yaml:"message"`
}

func (e Error) Error() string {
	return e.Message
}

func WrapInvalidArgumentErr(err error) error {
	return &Error{
		Code:    status.InvalidArgument,
		Message: err.Error(),
	}
}

func WrapInternalErr(err error) error {
	return &Error{
		Code:    status.Internal,
		Message: err.Error(),
	}
}

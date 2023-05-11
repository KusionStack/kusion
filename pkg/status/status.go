package status

import (
	"fmt"
)

type (
	Kind string
	Code string
)

const (
	Error   Kind = "ERROR"
	Warning Kind = "WARNING"
	Info    Kind = "INFO"
)

const (
	Unknown          Code = "UNKNOWN"
	Unavailable      Code = "UNAVAILABLE"
	Unimplemented    Code = "UNIMPLEMENTED"
	Canceled         Code = "CANCELED"
	InvalidArgument  Code = "INVALID_ARGUMENT"
	NotFound         Code = "NOTFOUND"
	AlreadyExists    Code = "ALREADY_EXISTS"
	PermissionDenied Code = "PERMISSION_DENIED"
	Internal         Code = "INTERNAL"
	Unauthenticated  Code = "UNAUTHENTICATED"
	IllegalManifest  Code = "ILLEGAL_MANIFEST"
)

type Status interface {
	Kind() Kind
	Code() Code
	Message() string
	String() string
}

type BaseStatus struct {
	kind    Kind
	code    Code
	message string
}

func (b *BaseStatus) Kind() Kind {
	return b.kind
}

func (b *BaseStatus) Code() Code {
	return b.code
}

func (b *BaseStatus) Message() string {
	return b.message
}

func (b *BaseStatus) String() string {
	return fmt.Sprintf("Kind: %s, Code: %s, Message: %s", b.kind, b.code, b.message)
}

func IsErr(s Status) bool {
	return s != nil && s.Kind() == Error
}

func NewBaseStatus(kind Kind, code Code, message string) *BaseStatus {
	return &BaseStatus{kind: kind, code: code, message: message}
}

func NewErrorStatus(err error) *BaseStatus {
	return &BaseStatus{kind: Error, code: Internal, message: err.Error()}
}

func NewErrorStatusWithCode(code Code, err error) *BaseStatus {
	return &BaseStatus{kind: Error, code: code, message: err.Error()}
}

func NewErrorStatusWithMsg(code Code, msg string) *BaseStatus {
	return &BaseStatus{kind: Error, code: code, message: msg}
}

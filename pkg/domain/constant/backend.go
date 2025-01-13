package constant

import "errors"

var (
	ErrInvalidBackendName = errors.New("backend name can only have alphanumeric characters and underscores with [a-zA-Z0-9_]")
	ErrEmptyBackendType   = errors.New("backend type is required")
	ErrInvalidBackendType = errors.New("backend type is should be one of the following: [local, oss, s3, google]")
)

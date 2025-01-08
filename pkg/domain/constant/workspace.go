package constant

import "errors"

var (
	ErrEmptyWorkspaceName   = errors.New("workspace must have a name")
	ErrInvalidWorkspaceName = errors.New("workspace name can only have alphanumeric characters and underscores with [a-zA-Z0-9_]")
	ErrEmptyOwners          = errors.New("workspace must have at least one owner")
	ErrEmptyBackendID       = errors.New("workspace must have a backend id")
)

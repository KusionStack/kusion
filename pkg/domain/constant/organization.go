package constant

import "errors"

// TODO: use v1.BackendType instead
// type BackendType string

// const (
// 	// SourceProviderTypeGithub represents github source provider type.
// 	BackendTypeOss   BackendType = "oss"
// 	BackendTypeLocal BackendType = "local"
// )

var (
	ErrOrgNil                  = errors.New("organization is nil")
	ErrOrgNameEmpty            = errors.New("organization must have a name")
	ErrOrgOwnerNil             = errors.New("org must have at least one owner")
	ErrWorkspaceNil            = errors.New("workspace is nil")
	ErrWorkspaceNameEmpty      = errors.New("workspace must have a name")
	ErrWorkspaceBackendNil     = errors.New("workspace must have a backend")
	ErrWorkspaceOwnerNil       = errors.New("workspace must have at least one owner")
	ErrBackendNil              = errors.New("backend is nil")
	ErrBackendNameEmpty        = errors.New("backend must have a name")
	ErrBackendTypeEmpty        = errors.New("backend must have a type")
	ErrDefaultBackendNotSet    = errors.New("default backend not set properly")
	ErrAppConfigHasNilStack    = errors.New("appConfig has nil stack")
	ErrInvalidOrganizationName = errors.New("organization name can only have alphanumeric characters and underscores with [a-zA-Z0-9_]")
	ErrInvalidOrganizationID   = errors.New("the organization ID should be a uuid")
)

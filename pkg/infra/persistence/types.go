package persistence

import (
	"errors"
)

var (
	ErrSourceModelNil                 = errors.New("source model can't be nil")
	ErrSystemConfigModelNil           = errors.New("system config model can't be nil")
	ErrStackModelNil                  = errors.New("stack model can't be nil")
	ErrProjectModelNil                = errors.New("project model can't be nil")
	ErrFailedToGetSourceProviderType  = errors.New("failed to parse source provider type")
	ErrFailedToGetSourceRemote        = errors.New("failed to parse source remote")
	ErrFailedToGetStackState          = errors.New("failed to parse stack state")
	ErrFailedToConvertSourceToEntity  = errors.New("failed to convert source model to entity")
	ErrFailedToConvertProjectToEntity = errors.New("failed to convert project model to entity")
	ErrFailedToConvertOrgToEntity     = errors.New("failed to convert org model to entity")
	ErrFailedToConvertBackendToEntity = errors.New("failed to convert backend model to entity")
	ErrOrganizationModelNil           = errors.New("organization model can't be nil")
	ErrWorkspaceModelNil              = errors.New("workspace model can't be nil")
	ErrBackendModelNil                = errors.New("backend model can't be nil")
)

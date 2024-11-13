package module

import (
	modulemanager "kusionstack.io/kusion/pkg/server/manager/module"
)

type Handler struct {
	moduleManager *modulemanager.ModuleManager
}

func NewHandler(
	moduleManager *modulemanager.ModuleManager,
) (*Handler, error) {
	return &Handler{
		moduleManager: moduleManager,
	}, nil
}

type ModuleRequestParams struct {
	ModuleName  string
	WorkspaceID uint
}

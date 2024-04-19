package workspace

import (
	workspacemanager "kusionstack.io/kusion/pkg/server/manager/workspace"
)

func NewHandler(
	workspaceManager *workspacemanager.WorkspaceManager,
) (*Handler, error) {
	return &Handler{
		workspaceManager: workspaceManager,
	}, nil
}

type Handler struct {
	workspaceManager *workspacemanager.WorkspaceManager
}

type WorkspaceRequestParams struct {
	WorkspaceID uint
}

package project

import (
	projectmanager "kusionstack.io/kusion/pkg/server/manager/project"
)

func NewHandler(
	projectManager *projectmanager.ProjectManager,
) (*Handler, error) {
	return &Handler{
		projectManager: projectManager,
	}, nil
}

type Handler struct {
	projectManager *projectmanager.ProjectManager
}

type ProjectRequestParams struct {
	ProjectID uint
}

package backend

import (
	backendmanager "kusionstack.io/kusion/pkg/server/manager/backend"
)

func NewHandler(
	backendManager *backendmanager.BackendManager,
) (*Handler, error) {
	return &Handler{
		backendManager: backendManager,
	}, nil
}

type Handler struct {
	backendManager *backendmanager.BackendManager
}

type BackendRequestParams struct {
	BackendID uint
}

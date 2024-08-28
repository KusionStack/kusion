package resource

import (
	resourcemanager "kusionstack.io/kusion/pkg/server/manager/resource"
)

func NewHandler(
	resourceManager *resourcemanager.ResourceManager,
) (*Handler, error) {
	return &Handler{
		resourceManager: resourceManager,
	}, nil
}

type Handler struct {
	resourceManager *resourcemanager.ResourceManager
}

type ResourceRequestParams struct {
	ResourceID uint
}

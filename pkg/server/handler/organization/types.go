package organization

import (
	organizationmanager "kusionstack.io/kusion/pkg/server/manager/organization"
)

func NewHandler(
	organizationManager *organizationmanager.OrganizationManager,
) (*Handler, error) {
	return &Handler{
		organizationManager: organizationManager,
	}, nil
}

type Handler struct {
	organizationManager *organizationmanager.OrganizationManager
}

type OrganizationRequestParams struct {
	OrganizationID uint
}

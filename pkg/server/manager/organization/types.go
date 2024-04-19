package organization

import (
	"errors"

	"kusionstack.io/kusion/pkg/domain/repository"
)

var (
	ErrGettingNonExistingOrganization  = errors.New("the organization does not exist")
	ErrUpdatingNonExistingOrganization = errors.New("the organization to update does not exist")
	ErrInvalidOrganizationID           = errors.New("the organization ID should be a uuid")
)

type OrganizationManager struct {
	organizationRepo repository.OrganizationRepository
}

func NewOrganizationManager(organizationRepo repository.OrganizationRepository) *OrganizationManager {
	return &OrganizationManager{
		organizationRepo: organizationRepo,
	}
}

package resource

import (
	"errors"

	"kusionstack.io/kusion/pkg/domain/repository"
)

var (
	ErrGettingNonExistingResource = errors.New("the resource does not exist")
	ErrInvalidResourceID          = errors.New("invalid resource ID")
)

type ResourceManager struct {
	resourceRepo repository.ResourceRepository
}

func NewResourceManager(resourceRepo repository.ResourceRepository) *ResourceManager {
	return &ResourceManager{
		resourceRepo: resourceRepo,
	}
}

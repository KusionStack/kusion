package backend

import (
	"errors"

	"kusionstack.io/kusion/pkg/domain/repository"
)

var (
	ErrGettingNonExistingBackend  = errors.New("the backend does not exist")
	ErrUpdatingNonExistingBackend = errors.New("the backend to update does not exist")
	ErrInvalidBackendID           = errors.New("the backend ID should be a uuid")
)

type BackendManager struct {
	backendRepo repository.BackendRepository
}

func NewBackendManager(backendRepo repository.BackendRepository) *BackendManager {
	return &BackendManager{
		backendRepo: backendRepo,
	}
}

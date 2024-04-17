package source

import (
	"errors"

	"kusionstack.io/kusion/pkg/domain/repository"
)

var (
	ErrGettingNonExistingSource  = errors.New("the source does not exist")
	ErrUpdatingNonExistingSource = errors.New("the source to update does not exist")
	ErrInvalidSourceID           = errors.New("the source ID should be a uuid")
)

type SourceManager struct {
	sourceRepo repository.SourceRepository
}

func NewSourceManager(sourceRepo repository.SourceRepository) *SourceManager {
	return &SourceManager{
		sourceRepo: sourceRepo,
	}
}

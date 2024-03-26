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

func NewHandler(
	sourceRepo repository.SourceRepository,
) (*Handler, error) {
	return &Handler{
		sourceRepo: sourceRepo,
	}, nil
}

type Handler struct {
	sourceRepo repository.SourceRepository
}

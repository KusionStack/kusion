package source

import (
	sourcemanager "kusionstack.io/kusion/pkg/server/manager/source"
)

func NewHandler(
	sourceManager *sourcemanager.SourceManager,
) (*Handler, error) {
	return &Handler{
		sourceManager: sourceManager,
	}, nil
}

type Handler struct {
	sourceManager *sourcemanager.SourceManager
}

type SourceRequestParams struct {
	SourceID uint
}

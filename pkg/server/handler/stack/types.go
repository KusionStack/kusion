package stack

import (
	stackmanager "kusionstack.io/kusion/pkg/server/manager/stack"
)

func NewHandler(
	stackManager *stackmanager.StackManager,
) (*Handler, error) {
	return &Handler{
		stackManager: stackManager,
	}, nil
}

type Handler struct {
	stackManager *stackmanager.StackManager
}

type StackRequestParams struct {
	StackID   uint
	Workspace string
	Format    string
	Detail    bool
	Dryrun    bool
}

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

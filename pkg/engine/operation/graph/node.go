package graph

import (
	"kusionstack.io/kusion/pkg/apis/status"
)

type baseNode struct {
	ID string
}

func NewBaseNode(id string) (*baseNode, status.Status) {
	if id == "" {
		return nil, status.NewErrorStatusWithMsg(status.InvalidArgument, "node id can not be nil")
	}
	return &baseNode{ID: id}, nil
}

func (b *baseNode) Hashcode() interface{} {
	return b.ID
}

func (b *baseNode) Name() string {
	return b.ID
}

type RootNode struct{}

func (r *RootNode) Hashcode() interface{} {
	return "RootNode"
}

func (r *RootNode) Name() string {
	return "root"
}

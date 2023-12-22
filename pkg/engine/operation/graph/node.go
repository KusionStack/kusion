package graph

import v1 "kusionstack.io/kusion/pkg/apis/status/v1"

type baseNode struct {
	ID string
}

func NewBaseNode(id string) (*baseNode, v1.Status) {
	if id == "" {
		return nil, v1.NewErrorStatusWithMsg(v1.InvalidArgument, "node id can not be nil")
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

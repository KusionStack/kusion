package meta

import (
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/backend"
)

// MetaOptions are the meta-options that are available on all or most commands.
type MetaOptions struct {
	// RefProject references the project for this CLI invocation.
	RefProject *v1.Project

	// RefStack referenced the stack for this CLI invocation.
	RefStack *v1.Stack

	// RefWorkspace referenced the workspace for this CLI invocation.
	RefWorkspace *v1.Workspace

	// Backend referenced the target storage backend for this CLI invocation.
	Backend backend.Backend
}

func (o *MetaOptions) GetRefProject() *v1.Project {
	return o.RefProject
}

func (o *MetaOptions) GetRefStack() *v1.Stack {
	return o.RefStack
}

func (o *MetaOptions) GetRefWorkspace() *v1.Workspace {
	return o.RefWorkspace
}

func (o *MetaOptions) GetBackend() backend.Backend {
	return o.Backend
}

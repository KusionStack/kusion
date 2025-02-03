package options

import (
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/backend"
)

type Meta interface {
	GetRefProject() *v1.Project
	GetRefStack() *v1.Stack
	GetRefWorkspace() *v1.Workspace
	GetBackend() backend.Backend
}

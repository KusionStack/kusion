package graph

import (
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/status"
)

type ExecutableNode interface {
	Execute(operation *opsmodels.Operation) status.Status
}

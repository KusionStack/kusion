package graph

import (
	"kusionstack.io/kusion/pkg/apis/status"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
)

type ExecutableNode interface {
	Execute(operation *opsmodels.Operation) status.Status
}

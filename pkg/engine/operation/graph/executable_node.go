package graph

import (
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/status"
)

type ExecutableNode interface {
	Execute(operation *models.Operation) status.Status
}

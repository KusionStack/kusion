package graph

import (
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation/models"
)

type ExecutableNode interface {
	Execute(operation *models.Operation) v1.Status
}

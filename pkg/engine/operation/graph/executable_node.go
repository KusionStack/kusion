package graph

import (
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
)

type ExecutableNode interface {
	Execute(operation *opsmodels.Operation) v1.Status
}

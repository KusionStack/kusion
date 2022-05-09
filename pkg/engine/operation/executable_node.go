package operation

import "kusionstack.io/kusion/pkg/status"

type ExecutableNode interface {
	Execute(operation Operation) status.Status
}

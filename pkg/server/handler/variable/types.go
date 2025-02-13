package variable

import "kusionstack.io/kusion/pkg/server/manager/variable"

type Handler struct {
	variableManager *variable.VariableManager
}

func NewHandler(
	variableManager *variable.VariableManager,
) (*Handler, error) {
	return &Handler{
		variableManager: variableManager,
	}, nil
}

type VariableRequestParams struct {
	VariableSetName string
	VariableName    string
}

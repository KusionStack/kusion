package variableset

import "kusionstack.io/kusion/pkg/server/manager/variableset"

type Handler struct {
	variableSetManager *variableset.VariableSetManager
}

func NewHandler(
	variableSetManager *variableset.VariableSetManager,
) (*Handler, error) {
	return &Handler{
		variableSetManager: variableSetManager,
	}, nil
}

type VariableSetRequestParams struct {
	VariableSetName string
}

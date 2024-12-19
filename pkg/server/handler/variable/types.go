package variable

import (
	"kusionstack.io/kusion/pkg/server/manager/variable"
	"kusionstack.io/kusion/pkg/server/manager/variablelabels"
)

type Handler struct {
	variableManager       *variable.VariableManager
	variableLabelsManager *variablelabels.VariableLabelsManager
}

func NewHandler(
	variableManager *variable.VariableManager,
	variableLabelsManager *variablelabels.VariableLabelsManager,
) (*Handler, error) {
	return &Handler{
		variableManager:       variableManager,
		variableLabelsManager: variableLabelsManager,
	}, nil
}

type VariableParams struct {
	Fqn string
}

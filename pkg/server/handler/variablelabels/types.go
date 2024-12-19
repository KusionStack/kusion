package variablelabels

import (
	"kusionstack.io/kusion/pkg/server/manager/variablelabels"
)

type Handler struct {
	variableLabelsManager *variablelabels.VariableLabelsManager
}

func NewHandler(
	variableLabelsManager *variablelabels.VariableLabelsManager,
) (*Handler, error) {
	return &Handler{
		variableLabelsManager: variableLabelsManager,
	}, nil
}

type VariableLabelsParams struct {
	Key string
}

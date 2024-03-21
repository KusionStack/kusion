package api

import (
	"fmt"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/state"
	"kusionstack.io/kusion/pkg/log"
)

type APIOptions struct {
	Operator     string
	Cluster      string
	IgnoreFields []string
}

func NewAPIOptions() APIOptions {
	apiOptions := APIOptions{
		// Operator:     "operator",
		// Cluster:      "cluster",
		// IgnoreFields: []string{},
	}
	return apiOptions
}

// The Preview function calculates the upcoming actions of each resource
// through the execution Kusion Engine, and you can customize the
// runtime of engine and the state storage through `runtime` and
// `storage` parameters.
//
// Example:
//
//	o := NewPreviewOptions()
//	stateStorage := &states.FileSystemState{
//	    Path: filepath.Join(o.WorkDir, states.KusionState)
//	}
//	kubernetesRuntime, err := runtime.NewKubernetesRuntime()
//	if err != nil {
//	    return err
//	}
//
//	changes, err := Preview(o, kubernetesRuntime, stateStorage,
//	    planResources, project, stack, os.Stdout)
//	if err != nil {
//	    return err
//	}
func Preview(
	o *APIOptions,
	storage state.Storage,
	planResources *apiv1.Intent,
	project *apiv1.Project,
	stack *apiv1.Stack,
) (*opsmodels.Changes, error) {
	log.Info("Start compute preview changes ...")

	// Construct the preview operation
	pc := &operation.PreviewOperation{
		Operation: opsmodels.Operation{
			OperationType: opsmodels.ApplyPreview,
			Stack:         stack,
			StateStorage:  storage,
			IgnoreFields:  o.IgnoreFields,
			ChangeOrder:   &opsmodels.ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*opsmodels.ChangeStep{}},
		},
	}

	log.Info("Start call pc.Preview() ...")

	// parse cluster in arguments
	rsp, s := pc.Preview(&operation.PreviewRequest{
		Request: opsmodels.Request{
			Project:  project,
			Stack:    stack,
			Operator: o.Operator,
			Intent:   planResources,
		},
	})
	if v1.IsErr(s) {
		return nil, fmt.Errorf("preview failed.\n%s", s.String())
	}

	return opsmodels.NewChanges(project, stack, rsp.Order), nil
}

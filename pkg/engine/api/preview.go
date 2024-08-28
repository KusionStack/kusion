package api

import (
	"fmt"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/engine/operation"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform"
	"kusionstack.io/kusion/pkg/infra/util/semaphore"
	"kusionstack.io/kusion/pkg/log"
)

type APIOptions struct {
	Operator      string
	Cluster       string
	IgnoreFields  []string
	DryRun        bool
	MaxConcurrent int
}

func NewAPIOptions() APIOptions {
	apiOptions := APIOptions{
		// Operator:     "operator",
		// Cluster:      "cluster",
		// IgnoreFields: []string{},
		DryRun:        false,
		MaxConcurrent: constant.MaxConcurrent,
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
	storage release.Storage,
	planResources *apiv1.Spec,
	priorResources *apiv1.State,
	proj *apiv1.Project,
	stack *apiv1.Stack,
) (*opsmodels.Changes, error) {
	log.Info("Start compute preview changes ...")

	// check and install terraform executable binary for
	// resources with the type of Terraform.
	tfInstaller := terraform.CLIInstaller{
		Intent: planResources,
	}
	if err := tfInstaller.CheckAndInstall(); err != nil {
		return nil, err
	}

	// construct the preview operation
	pc := &operation.PreviewOperation{
		Operation: opsmodels.Operation{
			OperationType:  opsmodels.ApplyPreview,
			Stack:          stack,
			ReleaseStorage: storage,
			IgnoreFields:   o.IgnoreFields,
			ChangeOrder:    &opsmodels.ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*opsmodels.ChangeStep{}},
			Sem:            semaphore.New(int64(o.MaxConcurrent)),
		},
	}

	log.Info("Start call pc.Preview() ...")

	// parse cluster in arguments
	rsp, s := pc.Preview(&operation.PreviewRequest{
		Request: opsmodels.Request{
			Project: proj,
			Stack:   stack,
		},
		Spec:  planResources,
		State: priorResources,
	})
	if v1.IsErr(s) {
		return nil, fmt.Errorf("preview failed.\n%s", s.String())
	}

	return opsmodels.NewChanges(proj, stack, rsp.Order), nil
}

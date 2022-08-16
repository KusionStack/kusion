package terraform

import (
	"context"
	"fmt"

	"github.com/imdario/mergo"
	"github.com/spf13/afero"
	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform/tfops"
	"kusionstack.io/kusion/pkg/status"
)

var _ runtime.Runtime = &TerraformRuntime{}

type TerraformRuntime struct {
	tfops.WorkspaceStore
}

func NewTerraformRuntime() (runtime.Runtime, error) {
	fs := afero.Afero{Fs: afero.NewOsFs()}
	ws, err := tfops.GetWorkspaceStore(fs)
	if err != nil {
		return nil, err
	}
	TFRuntime := &TerraformRuntime{ws}
	return TFRuntime, nil
}

// Apply terraform apply resource
func (t *TerraformRuntime) Apply(ctx context.Context, request *runtime.ApplyRequest) *runtime.ApplyResponse {
	planState := request.PlanResource
	w, ok := t.Store[planState.ResourceKey()]
	if !ok {
		err := t.Create(ctx, planState)
		if err != nil {
			return &runtime.ApplyResponse{Resource: nil, Status: status.NewErrorStatus(err)}
		}
		w = t.Store[planState.ResourceKey()]
	}

	// get terraform provider version
	providerAddr, err := w.GetProvider()
	if err != nil {
		return &runtime.ApplyResponse{Resource: nil, Status: status.NewErrorStatus(err)}
	}

	// terraform dry run merge state
	// TODO: terraform dry run apply,not only merge state
	if request.DryRun {
		prior := request.PriorResource.DeepCopy()
		if err := mergo.Merge(prior, planState, mergo.WithSliceDeepCopy, mergo.WithOverride); err != nil {
			return &runtime.ApplyResponse{Resource: nil, Status: status.NewErrorStatus(err)}
		}

		return &runtime.ApplyResponse{Resource: &models.Resource{
			ID:         planState.ID,
			Type:       planState.Type,
			Attributes: prior.Attributes,
			DependsOn:  planState.DependsOn,
			Extensions: planState.Extensions,
		}, Status: nil}
	}
	w.SetResource(planState)

	if err := w.WriteHCL(); err != nil {
		return &runtime.ApplyResponse{Resource: nil, Status: status.NewErrorStatus(err)}
	}

	tfstate, err := w.Apply(ctx)
	if err != nil {
		return &runtime.ApplyResponse{Resource: nil, Status: status.NewErrorStatus(err)}
	}

	r := tfops.ConvertTFState(tfstate, providerAddr)

	return &runtime.ApplyResponse{
		Resource: &models.Resource{
			ID:         r.ID,
			Type:       r.Type,
			Attributes: r.Attributes,
			DependsOn:  planState.DependsOn,
			Extensions: planState.Extensions,
		},
		Status: nil,
	}
}

// Read terraform show state
func (t *TerraformRuntime) Read(ctx context.Context, request *runtime.ReadRequest) *runtime.ReadResponse {
	priorState := request.PriorResource
	planState := request.PlanResource
	if priorState == nil {
		return &runtime.ReadResponse{Resource: nil, Status: nil}
	}
	var tfstate *tfops.TFState
	w, ok := t.Store[planState.ResourceKey()]
	if !ok {
		err := t.Create(ctx, planState)
		if err != nil {
			return &runtime.ReadResponse{Resource: nil, Status: status.NewErrorStatus(err)}
		}
		w = t.Store[priorState.ResourceKey()]
		if err := w.WriteTFState(priorState); err != nil {
			return &runtime.ReadResponse{Resource: nil, Status: status.NewErrorStatus(err)}
		}
	}

	tfstate, err := w.RefreshOnly(ctx)
	if err != nil {
		return &runtime.ReadResponse{Resource: nil, Status: status.NewErrorStatus(err)}
	}
	if tfstate == nil || tfstate.Values == nil {
		return &runtime.ReadResponse{Resource: nil, Status: nil}
	}

	// get terraform provider addr
	providerAddr, err := w.GetProvider()
	if err != nil {
		return &runtime.ReadResponse{Resource: nil, Status: status.NewErrorStatus(err)}
	}

	r := tfops.ConvertTFState(tfstate, providerAddr)
	return &runtime.ReadResponse{
		Resource: &models.Resource{
			ID:         r.ID,
			Type:       r.Type,
			Attributes: r.Attributes,
			DependsOn:  planState.DependsOn,
			Extensions: planState.Extensions,
		},
		Status: nil,
	}
}

// Delete terraform resource and remove workspace
func (t *TerraformRuntime) Delete(ctx context.Context, request *runtime.DeleteRequest) *runtime.DeleteResponse {
	w, ok := t.Store[request.Resource.ResourceKey()]
	if !ok {
		return &runtime.DeleteResponse{Status: status.NewErrorStatus(fmt.Errorf("%s terraform workspace not exist, cannot delete", request.Resource.ResourceKey()))}
	}
	if err := w.Destroy(ctx); err != nil {
		return &runtime.DeleteResponse{Status: status.NewErrorStatus(err)}
	}

	if err := t.Remove(ctx, request.Resource); err != nil {
		return &runtime.DeleteResponse{Status: status.NewErrorStatus(err)}
	}
	return &runtime.DeleteResponse{Status: nil}
}

// Watch terraform resource
func (t *TerraformRuntime) Watch(ctx context.Context, request *runtime.WatchRequest) *runtime.WatchResponse {
	return nil
}

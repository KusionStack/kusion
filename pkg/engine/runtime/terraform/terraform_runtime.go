package terraform

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/imdario/mergo"
	"github.com/spf13/afero"
	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform/tfops"
	"kusionstack.io/kusion/pkg/status"
)

var _ runtime.Runtime = &TerraformRuntime{}

type TerraformRuntime struct {
	tfops.WorkSpace
	mu *sync.Mutex
}

func NewTerraformRuntime() (runtime.Runtime, error) {
	fs := afero.Afero{Fs: afero.NewOsFs()}
	ws := tfops.NewWorkSpace(fs)
	TFRuntime := &TerraformRuntime{
		WorkSpace: *ws,
		mu:        &sync.Mutex{},
	}
	return TFRuntime, nil
}

// Apply terraform apply resource
func (t *TerraformRuntime) Apply(ctx context.Context, request *runtime.ApplyRequest) *runtime.ApplyResponse {
	planState := request.PlanResource
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

	t.mu.Lock()
	stackPath := request.Stack.GetPath()
	tfCacheDir := filepath.Join(stackPath, "."+planState.ResourceKey())
	t.WorkSpace.SetStackDir(stackPath)
	t.WorkSpace.SetCacheDir(tfCacheDir)
	t.WorkSpace.SetResource(planState)

	if err := t.WorkSpace.WriteHCL(); err != nil {
		return &runtime.ApplyResponse{Resource: nil, Status: status.NewErrorStatus(err)}
	}

	_, err := os.Stat(filepath.Join(tfCacheDir, tfops.HCLLOCKFILE))
	if err != nil {
		if os.IsNotExist(err) {
			if err := t.WorkSpace.InitWorkSpace(ctx); err != nil {
				return &runtime.ApplyResponse{Resource: nil, Status: status.NewErrorStatus(err)}
			}
		} else {
			return &runtime.ApplyResponse{Resource: nil, Status: status.NewErrorStatus(err)}
		}
	}

	tfstate, err := t.WorkSpace.Apply(ctx)
	if err != nil {
		return &runtime.ApplyResponse{Resource: nil, Status: status.NewErrorStatus(err)}
	}
	t.mu.Unlock()

	// get terraform provider version
	providerAddr, err := t.WorkSpace.GetProvider()
	if err != nil {
		return &runtime.ApplyResponse{Resource: nil, Status: status.NewErrorStatus(err)}
	}

	r := tfops.ConvertTFState(tfstate, providerAddr)

	return &runtime.ApplyResponse{
		Resource: &models.Resource{
			ID:         planState.ID,
			Type:       planState.Type,
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
	requestResource := request.PlanResource

	// When the operation is create or update, the requestResource is set to planResource,
	// when the operation is delete, planResource is nil, the requestResource is set to priorResource,
	// tf runtime uses requestResource to rebuild tfcache resources.
	if requestResource == nil {
		requestResource = request.PriorResource
	}
	if priorState == nil {
		return &runtime.ReadResponse{Resource: nil, Status: nil}
	}
	var tfstate *tfops.TFState

	t.mu.Lock()
	stackPath := request.Stack.GetPath()
	tfCacheDir := filepath.Join(stackPath, "."+requestResource.ResourceKey())
	t.WorkSpace.SetStackDir(stackPath)
	t.WorkSpace.SetCacheDir(tfCacheDir)
	t.WorkSpace.SetResource(requestResource)
	if err := t.WorkSpace.WriteHCL(); err != nil {
		return &runtime.ReadResponse{Resource: nil, Status: status.NewErrorStatus(err)}
	}
	_, err := os.Stat(filepath.Join(tfCacheDir, tfops.HCLLOCKFILE))
	if err != nil {
		if os.IsNotExist(err) {
			if err := t.WorkSpace.InitWorkSpace(ctx); err != nil {
				return &runtime.ReadResponse{Resource: nil, Status: status.NewErrorStatus(err)}
			}
		} else {
			return &runtime.ReadResponse{Resource: nil, Status: status.NewErrorStatus(err)}
		}
	}

	// priorState overwirte tfstate in workspace
	if err := t.WorkSpace.WriteTFState(priorState); err != nil {
		return &runtime.ReadResponse{Resource: nil, Status: status.NewErrorStatus(err)}
	}

	tfstate, err = t.WorkSpace.RefreshOnly(ctx)
	if err != nil {
		return &runtime.ReadResponse{Resource: nil, Status: status.NewErrorStatus(err)}
	}
	t.mu.Unlock()
	if tfstate == nil || tfstate.Values == nil {
		return &runtime.ReadResponse{Resource: nil, Status: nil}
	}

	// get terraform provider addr
	providerAddr, err := t.WorkSpace.GetProvider()
	if err != nil {
		return &runtime.ReadResponse{Resource: nil, Status: status.NewErrorStatus(err)}
	}

	r := tfops.ConvertTFState(tfstate, providerAddr)
	return &runtime.ReadResponse{
		Resource: &models.Resource{
			ID:         requestResource.ID,
			Type:       requestResource.Type,
			Attributes: r.Attributes,
			DependsOn:  requestResource.DependsOn,
			Extensions: requestResource.Extensions,
		},
		Status: nil,
	}
}

// Delete terraform resource and remove workspace
func (t *TerraformRuntime) Delete(ctx context.Context, request *runtime.DeleteRequest) *runtime.DeleteResponse {
	stackPath := request.Stack.GetPath()
	tfCacheDir := filepath.Join(stackPath, "."+request.Resource.ResourceKey())
	defer os.RemoveAll(tfCacheDir)
	t.mu.Lock()
	t.WorkSpace.SetStackDir(stackPath)
	t.WorkSpace.SetCacheDir(tfCacheDir)
	t.WorkSpace.SetResource(request.Resource)
	if err := t.WorkSpace.Destroy(ctx); err != nil {
		return &runtime.DeleteResponse{Status: status.NewErrorStatus(err)}
	}
	t.mu.Unlock()

	return &runtime.DeleteResponse{Status: nil}
}

// Watch terraform resource
func (t *TerraformRuntime) Watch(ctx context.Context, request *runtime.WatchRequest) *runtime.WatchResponse {
	return nil
}

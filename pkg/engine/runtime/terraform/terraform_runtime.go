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
	"kusionstack.io/kusion/pkg/log"
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
	defer t.mu.Unlock()
	stackPath := request.Stack.GetPath()
	tfCacheDir := filepath.Join(stackPath, "."+planState.ResourceKey())
	t.WorkSpace.SetStackDir(stackPath)
	t.WorkSpace.SetCacheDir(tfCacheDir)
	t.WorkSpace.SetResource(planState)

	if err := t.WorkSpace.WriteHCL(); err != nil {
		return &runtime.ApplyResponse{Resource: nil, Status: status.NewErrorStatus(err)}
	}

	_, err := os.Stat(filepath.Join(tfCacheDir, tfops.LockHCLFile))
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
	priorResource := request.PriorResource
	requestResource := request.PlanResource

	// When the operation is create or update, the requestResource is set to planResource,
	// when the operation is delete, planResource is nil, the requestResource is set to priorResource,
	// tf runtime uses requestResource to rebuild tfcache resources.
	if requestResource == nil && priorResource != nil {
		// requestResource is nil representing that this is a Delete action.
		// We only need to refresh the tf.state files and return the latest resources state in this method.
		// Most fields in attributes in resources aren't necessary for the command `terraform apply -refresh-only` and will make errors
		// if fields copied from kusion_state.json but read-only in main.tf.json
		requestResource = &models.Resource{
			ID:         priorResource.ID,
			Type:       priorResource.Type,
			Attributes: nil,
			DependsOn:  priorResource.DependsOn,
			Extensions: priorResource.Extensions,
		}
	}
	if priorResource == nil {
		return &runtime.ReadResponse{Resource: nil, Status: nil}
	}
	var tfstate *tfops.TFState

	t.mu.Lock()
	defer t.mu.Unlock()
	stackPath := request.Stack.GetPath()
	tfCacheDir := filepath.Join(stackPath, "."+requestResource.ResourceKey())
	t.WorkSpace.SetStackDir(stackPath)
	t.WorkSpace.SetCacheDir(tfCacheDir)
	t.WorkSpace.SetResource(requestResource)

	if err := t.WorkSpace.WriteHCL(); err != nil {
		return &runtime.ReadResponse{Resource: nil, Status: status.NewErrorStatus(err)}
	}
	_, err := os.Stat(filepath.Join(tfCacheDir, tfops.LockHCLFile))
	if err != nil {
		if os.IsNotExist(err) {
			if err := t.WorkSpace.InitWorkSpace(ctx); err != nil {
				return &runtime.ReadResponse{Resource: nil, Status: status.NewErrorStatus(err)}
			}
		} else {
			return &runtime.ReadResponse{Resource: nil, Status: status.NewErrorStatus(err)}
		}
	}

	// priorResource overwrite tfstate in workspace
	if err = t.WorkSpace.WriteTFState(priorResource); err != nil {
		return &runtime.ReadResponse{Resource: nil, Status: status.NewErrorStatus(err)}
	}

	tfstate, err = t.WorkSpace.RefreshOnly(ctx)
	if err != nil {
		return &runtime.ReadResponse{Resource: nil, Status: status.NewErrorStatus(err)}
	}

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

func (t *TerraformRuntime) Import(ctx context.Context, request *runtime.ImportRequest) *runtime.ImportResponse {
	// TODO change to terraform cli import
	log.Info("skip import TF resource:%s", request.PlanResource.ID)
	return nil
}

// Delete terraform resource and remove workspace
func (t *TerraformRuntime) Delete(ctx context.Context, request *runtime.DeleteRequest) (res *runtime.DeleteResponse) {
	stackPath := request.Stack.GetPath()
	tfCacheDir := filepath.Join(stackPath, "."+request.Resource.ResourceKey())
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			res = &runtime.DeleteResponse{Status: status.NewErrorStatus(err)}
		}
	}(tfCacheDir)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.WorkSpace.SetStackDir(stackPath)
	t.WorkSpace.SetCacheDir(tfCacheDir)
	t.WorkSpace.SetResource(request.Resource)
	if err := t.WorkSpace.Destroy(ctx); err != nil {
		return &runtime.DeleteResponse{Status: status.NewErrorStatus(err)}
	}

	return &runtime.DeleteResponse{Status: nil}
}

// Watch terraform resource
func (t *TerraformRuntime) Watch(ctx context.Context, request *runtime.WatchRequest) *runtime.WatchResponse {
	return nil
}

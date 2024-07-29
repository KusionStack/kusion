package terraform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform/tfops"
	"kusionstack.io/kusion/pkg/log"
)

var _ runtime.Runtime = &Runtime{}

// tfEvents is used to record the operation events of the Terraform
// resources into the related channels for watching.
var tfEvents = cache.New(cache.NoExpiration, cache.NoExpiration)

type Runtime struct {
	mutex   *sync.Mutex
	context apiv1.GenericConfig
}

func NewTerraformRuntime(spec apiv1.Spec) (runtime.Runtime, error) {
	TFRuntime := &Runtime{
		mutex:   &sync.Mutex{},
		context: spec.Context,
	}
	return TFRuntime, nil
}

// Apply Terraform resource
func (t *Runtime) Apply(ctx context.Context, request *runtime.ApplyRequest) *runtime.ApplyResponse {
	plan := request.PlanResource
	stackPath := request.Stack.Path
	key := plan.ResourceKey()
	tfCacheDir := buildTFCacheDir(stackPath, key)
	ws := tfops.NewWorkSpace(plan, stackPath, tfCacheDir, t.mutex, t.context)

	if err := ws.WriteHCL(); err != nil {
		return &runtime.ApplyResponse{Resource: nil, Status: v1.NewErrorStatus(err)}
	}

	_, err := os.Stat(filepath.Join(tfCacheDir, tfops.LockHCLFile))
	if err != nil {
		if os.IsNotExist(err) {
			if err := ws.InitWorkSpace(ctx); err != nil {
				return &runtime.ApplyResponse{Resource: nil, Status: v1.NewErrorStatus(err)}
			}
		} else {
			return &runtime.ApplyResponse{Resource: nil, Status: v1.NewErrorStatus(err)}
		}
	}

	// dry run by terraform plan
	if request.DryRun {
		pr, err := ws.Plan(ctx)
		if err != nil {
			return &runtime.ApplyResponse{Resource: nil, Status: v1.NewErrorStatus(err)}
		}
		module := pr.PlannedValues.RootModule
		if len(module.Resources) == 0 {
			log.Debugf("no resource found in terraform plan file")
			return &runtime.ApplyResponse{Resource: &apiv1.Resource{}, Status: nil}
		}

		return &runtime.ApplyResponse{
			Resource: &apiv1.Resource{
				ID:         plan.ID,
				Type:       plan.Type,
				Attributes: module.Resources[0].AttributeValues,
				DependsOn:  plan.DependsOn,
				Extensions: plan.Extensions,
			},
			Status: nil,
		}
	}

	var tfstate *tfops.StateRepresentation
	var providerAddr string

	// Extract the watch channel from the context.
	watchCh, _ := ctx.Value(engine.WatchChannel).(chan string)
	if watchCh != nil {
		// Apply while watching the resource.
		errCh := make(chan error)

		// Start applying the resource.
		go func() {
			tfstate, err = ws.Apply(ctx)
			if err != nil {
				errCh <- err
			}

			providerAddr, err = ws.GetProvider()
			errCh <- err
		}()

		// Prepare the event channel and send the resource ID to watch channel.
		log.Infof("Started to watch %s with the type of %s", key, plan.Type)
		eventCh := make(chan runtime.TFEvent)

		// Prevent concurrent operations on resources with the same ID.
		if _, ok := tfEvents.Get(key); ok {
			err = fmt.Errorf("failed to initiate the event channel for watching terraform resource %s as: conflict resource ID", key)
			log.Error(err)
			return &runtime.ApplyResponse{Resource: nil, Status: v1.NewErrorStatus(err)}
		}
		tfEvents.Set(key, eventCh, cache.NoExpiration)
		watchCh <- key

		// Wait for the apply to be finished.
		shouldBreak := false
		for !shouldBreak {
			select {
			case err = <-errCh:
				if err != nil {
					eventCh <- runtime.TFFailed
					return &runtime.ApplyResponse{Resource: nil, Status: v1.NewErrorStatus(err)}
				}
				eventCh <- runtime.TFSucceeded
				shouldBreak = true
			default:
				eventCh <- runtime.TFApplying
				time.Sleep(time.Second * 1)
			}
		}
	} else {
		// Apply without watching.
		tfstate, err = ws.Apply(ctx)
		if err != nil {
			return &runtime.ApplyResponse{Resource: nil, Status: v1.NewErrorStatus(err)}
		}

		// get terraform provider version
		providerAddr, err = ws.GetProvider()
		if err != nil {
			return &runtime.ApplyResponse{Resource: nil, Status: v1.NewErrorStatus(err)}
		}
	}

	r := tfops.ConvertTFState(tfstate, providerAddr)

	return &runtime.ApplyResponse{
		Resource: &apiv1.Resource{
			ID:         plan.ID,
			Type:       plan.Type,
			Attributes: r.Attributes,
			DependsOn:  plan.DependsOn,
			Extensions: plan.Extensions,
		},
		Status: nil,
	}
}

func buildTFCacheDir(stackPath string, key string) string {
	// replace ':' with '_' to comply with Windows directory naming conventions.
	return filepath.Join(stackPath, "."+strings.ReplaceAll(key, ":", "_"))
}

// Read terraform show state
func (t *Runtime) Read(ctx context.Context, request *runtime.ReadRequest) *runtime.ReadResponse {
	priorResource := request.PriorResource
	planResource := request.PlanResource

	if priorResource == nil && planResource == nil {
		return &runtime.ReadResponse{Resource: nil, Status: nil}
	}

	// when the operation is delete, planResource is nil, the planResource is set to priorResource,
	// tf runtime uses planResource to rebuild tfcache resources.
	if planResource == nil {
		// planResource is nil representing that this is a Delete action.
		// We only need to refresh the tf.state files and return the latest resources state in this method.
		// Most fields in the `attributes` field of resource aren't necessary for the command `terraform apply -refresh-only`.
		// These fields will cause errors if they are copied from kusion_state.json but read-only in main.tf.json
		planResource = &apiv1.Resource{
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

	var tfState *tfops.StateRepresentation
	stackPath := request.Stack.Path
	tfCacheDir := buildTFCacheDir(stackPath, planResource.ResourceKey())

	ws := tfops.NewWorkSpace(planResource, stackPath, tfCacheDir, t.mutex, t.context)
	if err := ws.WriteHCL(); err != nil {
		return &runtime.ReadResponse{Resource: nil, Status: v1.NewErrorStatus(err)}
	}
	_, err := os.Stat(filepath.Join(tfCacheDir, tfops.LockHCLFile))
	if err != nil {
		if os.IsNotExist(err) {
			if err := ws.InitWorkSpace(ctx); err != nil {
				return &runtime.ReadResponse{Resource: nil, Status: v1.NewErrorStatus(err)}
			}
		} else {
			return &runtime.ReadResponse{Resource: nil, Status: v1.NewErrorStatus(err)}
		}
	}

	importID, ok := planResource.Extensions[tfops.ImportIDKey].(string)
	if ok && importID != "" {
		if err = ws.ImportResource(ctx, importID); err != nil {
			return &runtime.ReadResponse{Resource: nil, Status: v1.NewErrorStatus(err)}
		} else {
			// read resource from tfstate
			tfState, err = ws.ShowState(ctx)
			if err != nil {
				return &runtime.ReadResponse{Resource: nil, Status: v1.NewErrorStatus(err)}
			}
			// get terraform provider version
			providerAddr, err := ws.GetProvider()
			if err != nil {
				return &runtime.ReadResponse{Resource: nil, Status: v1.NewErrorStatus(err)}
			}
			r := tfops.ConvertTFState(tfState, providerAddr)
			return &runtime.ReadResponse{
				Resource: &apiv1.Resource{
					ID:         planResource.ID,
					Type:       planResource.Type,
					Attributes: r.Attributes,
					DependsOn:  planResource.DependsOn,
					Extensions: planResource.Extensions,
				}, Status: nil,
			}
		}
	} else if err = ws.WriteTFState(priorResource); err != nil {
		// priorResource overwrite tfState in workspace
		return &runtime.ReadResponse{Resource: nil, Status: v1.NewErrorStatus(err)}
	}

	tfState, err = ws.RefreshOnly(ctx)
	if err != nil {
		return &runtime.ReadResponse{Resource: nil, Status: v1.NewErrorStatus(err)}
	}

	if tfState == nil || tfState.Values == nil {
		return &runtime.ReadResponse{Resource: nil, Status: nil}
	}

	// get terraform provider addr
	providerAddr, err := ws.GetProvider()
	if err != nil {
		return &runtime.ReadResponse{Resource: nil, Status: v1.NewErrorStatus(err)}
	}

	r := tfops.ConvertTFState(tfState, providerAddr)
	return &runtime.ReadResponse{
		Resource: &apiv1.Resource{
			ID:         planResource.ID,
			Type:       planResource.Type,
			Attributes: r.Attributes,
			DependsOn:  planResource.DependsOn,
			Extensions: planResource.Extensions,
		},
		Status: nil,
	}
}

func (t *Runtime) Import(ctx context.Context, request *runtime.ImportRequest) *runtime.ImportResponse {
	response := t.Read(ctx, &runtime.ReadRequest{
		PlanResource: request.PlanResource,
		Stack:        request.Stack,
	})

	if v1.IsErr(response.Status) {
		return &runtime.ImportResponse{
			Resource: nil,
			Status:   response.Status,
		}
	}

	return &runtime.ImportResponse{
		Resource: response.Resource,
		Status:   nil,
	}
}

// Delete terraform resource and remove workspace
func (t *Runtime) Delete(ctx context.Context, request *runtime.DeleteRequest) (res *runtime.DeleteResponse) {
	stackPath := request.Stack.Path
	tfCacheDir := buildTFCacheDir(stackPath, request.Resource.ResourceKey())

	ws := tfops.NewWorkSpace(request.Resource, stackPath, tfCacheDir, t.mutex, t.context)
	if err := ws.Destroy(ctx); err != nil {
		return &runtime.DeleteResponse{Status: v1.NewErrorStatus(err)}
	}

	// delete tf directory after destroy operation is success
	err := os.RemoveAll(tfCacheDir)
	if err != nil {
		return &runtime.DeleteResponse{Status: v1.NewErrorStatus(err)}
	}
	return &runtime.DeleteResponse{Status: nil}
}

// Watch terraform resource
func (t *Runtime) Watch(ctx context.Context, request *runtime.WatchRequest) *runtime.WatchResponse {
	// Get the event channel.
	id := request.Resource.ResourceKey()
	eventCh, ok := tfEvents.Get(id)
	if !ok {
		return &runtime.WatchResponse{Status: v1.NewErrorStatus(fmt.Errorf("failed to get the event channel for %s", id))}
	}

	return &runtime.WatchResponse{
		Watchers: &runtime.SequentialWatchers{
			IDs:       []string{id},
			TFWatcher: eventCh.(chan runtime.TFEvent),
		},
	}
}

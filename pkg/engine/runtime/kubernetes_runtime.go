package runtime

import (
	"context"
	"errors"

	jsonpatch "github.com/evanphx/json-patch"
	yamlv2 "gopkg.in/yaml.v2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/status"
	"kusionstack.io/kusion/pkg/util/json"
	"kusionstack.io/kusion/pkg/util/kube/config"
)

var _ Runtime = (*KubernetesRuntime)(nil)

type KubernetesRuntime struct {
	dyn    dynamic.Interface
	mapper meta.RESTMapper
}

// NewKubernetesRuntime create a new KubernetesRuntime
func NewKubernetesRuntime() (Runtime, error) {
	dyn, mapper, err := getKubernetesClient()
	if err != nil {
		return nil, err
	}

	return &KubernetesRuntime{
		dyn:    dyn,
		mapper: mapper,
	}, nil
}

// Apply kubernetes Resource by client-go
func (k *KubernetesRuntime) Apply(ctx context.Context, request *ApplyRequest) *ApplyResponse {
	planState := request.PlanResource
	priorState := request.PriorResource

	// Don`t consider delete case, so plan state must be not empty
	if planState == nil {
		return &ApplyResponse{nil, status.NewErrorStatus(errors.New("plan state is nil"))}
	}

	// Get kubernetes Resource interface from plan state
	planObj, resource, err := k.buildKubernetesResourceByState(planState)
	if err != nil {
		return &ApplyResponse{nil, status.NewErrorStatus(err)}
	}

	// Get live state
	response := k.Read(ctx, &ReadRequest{PlanResource: planState})
	if status.IsErr(response.Status) {
		return &ApplyResponse{nil, response.Status}
	}
	liveState := response.Resource

	// Original equals to last-applied from annotation, kusion store it in kusion_state.json
	original := ""
	if priorState != nil {
		original = json.MustMarshal2String(priorState.Attributes)
	}
	// Modified equals to input content
	modified := json.MustMarshal2String(planState.Attributes)
	// Current equals to live manifest
	current := ""
	if liveState != nil {
		current = json.MustMarshal2String(liveState.Attributes)
	}

	// Create patch body
	patchBody, err := jsonmergepatch.CreateThreeWayJSONMergePatch([]byte(original), []byte(modified), []byte(current))
	if err != nil {
		return &ApplyResponse{nil, status.NewErrorStatus(err)}
	}

	// Final result, dry-run to diff, otherwise to save in states
	var res *unstructured.Unstructured
	if request.DryRun {
		tryClientSide := true
		if liveState == nil {
			// Try ServerSideDryRun first
			createOptions := metav1.CreateOptions{
				DryRun: []string{metav1.DryRunAll},
			}
			if createdObj, err := resource.Create(ctx, planObj, createOptions); err == nil {
				res = createdObj
				// ServerSideDryRun success, not need to try ClientSideDryRun
				tryClientSide = false
			} else {
				log.Errorf("ServerSideDryRun create failed, err: %v", err)
			}
		} else {
			// Try ServerSideDryRun first
			patchOptions := metav1.PatchOptions{
				DryRun: []string{metav1.DryRunAll},
			}
			if patchedObj, err := resource.Patch(ctx, planObj.GetName(), types.MergePatchType, patchBody, patchOptions); err == nil {
				res = patchedObj
				tryClientSide = false
			} else {
				log.Errorf("ServerSideDryRun Patch failed, err: %v", err)
			}
		}
		if tryClientSide {
			// Fall back to ClientSideDryRun
			mergedPatch, err := jsonpatch.MergePatch([]byte(current), patchBody)
			if err != nil {
				return &ApplyResponse{nil, status.NewErrorStatus(err)}
			}
			res = &unstructured.Unstructured{}
			if err = res.UnmarshalJSON(mergedPatch); err != nil {
				return &ApplyResponse{nil, status.NewErrorStatus(err)}
			}
		}
	} else {
		if liveState == nil {
			// LiveState is nil, fall back to create planObj
			_, err = resource.Create(ctx, planObj, metav1.CreateOptions{})
		} else {
			// LiveState isn't nil, continue to patch liveObj
			_, err = resource.Patch(ctx, planObj.GetName(), types.MergePatchType, patchBody, metav1.PatchOptions{FieldManager: "kusion"})
		}
		if err != nil {
			return &ApplyResponse{nil, status.NewErrorStatus(err)}
		}
		// Save modified
		res = planObj
	}

	return &ApplyResponse{&models.Resource{
		ID:         planState.ResourceKey(),
		Type:       planState.Type,
		Attributes: res.Object,
		DependsOn:  planState.DependsOn,
		Extensions: planState.Extensions,
	}, nil}
}

// Read kubernetes Resource by client-go
func (k *KubernetesRuntime) Read(ctx context.Context, request *ReadRequest) *ReadResponse {
	requestResource := request.PlanResource
	if requestResource == nil {
		requestResource = request.PriorResource
	}
	// Validate
	if requestResource == nil {
		return &ReadResponse{nil, status.NewErrorStatus(errors.New("requestResource is nil"))}
	}

	// Get resource by attribute
	obj, resource, err := k.buildKubernetesResourceByState(requestResource)
	if err != nil {
		// Ignore no match error, cause target apiVersion or kind is not installed yet
		if meta.IsNoMatchError(err) {
			log.Infof("%v, ignore", err)
			return &ReadResponse{nil, nil}
		}
		return &ReadResponse{nil, status.NewErrorStatus(err)}
	}

	// Read resource
	v, err := resource.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Infof("%s not found, ignore", requestResource.ResourceKey())
			return &ReadResponse{nil, nil}
		}
		return &ReadResponse{nil, status.NewErrorStatus(err)}
	}

	return &ReadResponse{&models.Resource{
		ID:         requestResource.ResourceKey(),
		Type:       requestResource.Type,
		Attributes: v.Object,
		DependsOn:  requestResource.DependsOn,
		Extensions: requestResource.Extensions,
	}, nil}
}

// Delete kubernetes Resource by client-go
func (k *KubernetesRuntime) Delete(ctx context.Context, request *DeleteRequest) *DeleteResponse {
	requestResource := request.Resource
	// Validate
	if requestResource == nil {
		return &DeleteResponse{status.NewErrorStatus(errors.New("requestResource is nil"))}
	}

	// Get Resource by attribute
	obj, resource, err := k.buildKubernetesResourceByState(requestResource)
	if err != nil {
		return &DeleteResponse{status.NewErrorStatus(err)}
	}

	// Delete Resource
	err = resource.Delete(ctx, obj.GetName(), metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Infof("%s not found, ignore", requestResource.ResourceKey())
			return &DeleteResponse{nil}
		}
		return &DeleteResponse{status.NewErrorStatus(err)}
	}

	return &DeleteResponse{nil}
}

// Watch kubernetes resource by client-go
func (k *KubernetesRuntime) Watch(ctx context.Context, request *WatchRequest) *WatchResponse {
	panic("need implement")
}

// getKubernetesClient get kubernetes client
func getKubernetesClient() (dynamic.Interface, meta.RESTMapper, error) {
	// build config
	cfg, err := clientcmd.BuildConfigFromFlags("", config.GetKubeConfig())
	if err != nil {
		return nil, nil, err
	}

	// DynamicRESTMapper can discover resource types at runtime dynamically
	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		return nil, nil, err
	}

	// Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	return dyn, mapper, nil
}

// buildKubernetesResourceByState get resource by attribute
func (k *KubernetesRuntime) buildKubernetesResourceByState(resourceState *models.Resource) (*unstructured.Unstructured, dynamic.ResourceInterface, error) {
	// Convert interface{} to unstructured
	rYaml, err := yamlv2.Marshal(resourceState.Attributes)
	if err != nil {
		return nil, nil, err
	}

	obj, gvk, err := convertString2Unstructured(rYaml)
	if err != nil {
		return nil, nil, err
	}

	// Get resource by unstructured
	var resource dynamic.ResourceInterface

	resource, err = buildKubernetesResourceByUnstructured(k.dyn, k.mapper, obj, gvk)
	if err != nil {
		return nil, nil, err
	}

	return obj, resource, nil
}

// buildKubernetesResourceByUnstructured get resource by unstructured object
func buildKubernetesResourceByUnstructured(dyn dynamic.Interface, mapper meta.RESTMapper,
	obj *unstructured.Unstructured, gvk *schema.GroupVersionKind,
) (dynamic.ResourceInterface, error) {
	// Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	// Obtain REST interface for the GVR
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dr = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = dyn.Resource(mapping.Resource)
	}

	return dr, nil
}

// convertString2Unstructured convert string to unstructured object
func convertString2Unstructured(yamlContent []byte) (*unstructured.Unstructured, *schema.GroupVersionKind, error) {
	// Decode YAML manifest into unstructured.Unstructured
	decUnstructured := k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}

	_, gvk, err := decUnstructured.Decode(yamlContent, nil, obj)
	if err != nil {
		return nil, nil, err
	}

	return obj, gvk, nil
}

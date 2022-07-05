package runtime

import (
	"context"
	"errors"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"

	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/status"
	"kusionstack.io/kusion/pkg/util/json"
	"kusionstack.io/kusion/pkg/util/kube/config"
	"kusionstack.io/kusion/pkg/util/yaml"
)

var _ Runtime = (*KubernetesRuntime)(nil)

type KubernetesRuntime struct {
	dyn    dynamic.Interface
	mapper *restmapper.DeferredDiscoveryRESTMapper
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
	response := k.Read(ctx, &ReadRequest{planState})
	liveState := response.Resource
	s := response.Status
	if status.IsErr(s) {
		return &ApplyResponse{nil, s}
	}

	// LiveState is nil, fall back to create planObj directly
	if liveState == nil {
		if _, err = resource.Create(ctx, planObj, metav1.CreateOptions{}); err != nil {
			return &ApplyResponse{nil, status.NewErrorStatus(err)}
		}
	} else {
		// Original equals to last-applied from annotation, kusion store it in kusion_state.json
		original := ""
		if priorState != nil {
			original = json.MustMarshal2String(priorState.Attributes)
		}
		// Modified equals input content
		modified := json.MustMarshal2String(planState.Attributes)
		// Current equals live manifest
		current := json.MustMarshal2String(liveState.Attributes)
		// 3-way json merge patch
		patch, err := jsonmergepatch.CreateThreeWayJSONMergePatch([]byte(original), []byte(modified), []byte(current))
		if err != nil {
			return &ApplyResponse{nil, status.NewErrorStatus(err)}
		}
		// Apply patch
		if _, err = resource.Patch(ctx, planObj.GetName(), types.MergePatchType, patch, metav1.PatchOptions{FieldManager: "kusion"}); err != nil {
			return &ApplyResponse{nil, status.NewErrorStatus(err)}
		}
	}

	return &ApplyResponse{&models.Resource{
		ID:         planState.ResourceKey(),
		Attributes: planObj.Object,
		DependsOn:  planState.DependsOn,
	}, nil}
}

// Read kubernetes Resource by client-go
func (k *KubernetesRuntime) Read(ctx context.Context, request *ReadRequest) *ReadResponse {
	requestResource := request.Resource
	// Validate
	if requestResource == nil {
		return &ReadResponse{nil, status.NewErrorStatus(errors.New("requestResource is nil"))}
	}

	// Get resource by attribute
	obj, resource, err := k.buildKubernetesResourceByState(requestResource)
	if err != nil {
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
		Attributes: v.Object,
		DependsOn:  requestResource.DependsOn,
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
func getKubernetesClient() (dynamic.Interface, *restmapper.DeferredDiscoveryRESTMapper, error) {
	// build config
	cfg, err := clientcmd.BuildConfigFromFlags("", config.GetKubeConfig())
	if err != nil {
		return nil, nil, err
	}

	// Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

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
	attribute := resourceState.Attributes
	rYaml := yaml.MergeToOneYAML(attribute)

	obj, gvk, err := convertString2Unstructured([]byte(rYaml))
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
func buildKubernetesResourceByUnstructured(dyn dynamic.Interface, mapper *restmapper.DeferredDiscoveryRESTMapper,
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

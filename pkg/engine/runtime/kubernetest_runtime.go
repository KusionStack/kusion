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

// Apply kubernetes resource by client-go
func (k *KubernetesRuntime) Apply(ctx context.Context, priorState, planState *models.Resource) (*models.Resource, status.Status) {
	// Don`t consider delete case, so plan state must be not empty
	if planState == nil {
		return nil, status.NewErrorStatus(errors.New("plan state is nil"))
	}

	// Get kubernetes resource interface from plan state
	planObj, resource, err := k.buildKubernetesResourceByState(planState)
	if err != nil {
		return nil, status.NewErrorStatus(err)
	}

	// original equals to last-applied from annotation, kusion store it in kusion_state.json
	original := yaml.MergeToOneYAML(priorState.Attributes)
	// modified equals input content
	modified := yaml.MergeToOneYAML(planState.Attributes)
	// get live state
	liveState, err := resource.Get(ctx, planObj.GetName(), metav1.GetOptions{})
	if err != nil {
		return nil, status.NewErrorStatus(err)
	}
	// current equals live manifest
	current := yaml.MergeToOneYAML(liveState.Object)
	// 3-way json merge patch
	patch, err := jsonmergepatch.CreateThreeWayJSONMergePatch([]byte(original), []byte(modified), []byte(current))
	if err != nil {
		return nil, status.NewErrorStatus(err)
	}
	// apply patch
	patchedObj, err := resource.Patch(ctx, planObj.GetName(), types.MergePatchType, patch, metav1.PatchOptions{
		FieldManager: "kusion",
	})
	if err != nil {
		return nil, status.NewErrorStatus(err)
	}

	return &models.Resource{
		ID:         planState.ResourceKey(),
		Attributes: patchedObj.Object,
		DependsOn:  planState.DependsOn,
	}, nil
}

// Read kubernetes resource by client-go
func (k *KubernetesRuntime) Read(ctx context.Context, resourceState *models.Resource) (*models.Resource, status.Status) {
	// Validate
	if resourceState == nil {
		return nil, status.NewErrorStatus(errors.New("resourceState is nil"))
	}

	// Get resource by attribute
	obj, resource, err := k.buildKubernetesResourceByState(resourceState)
	if err != nil {
		return nil, status.NewErrorStatus(err)
	}

	// Read resource
	v, err := resource.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		return nil, status.NewErrorStatus(err)
	}

	return &models.Resource{
		ID:         resourceState.ResourceKey(),
		Attributes: v.Object,
		DependsOn:  resourceState.DependsOn,
	}, nil
}

// Delete kubernetes resource by client-go
func (k *KubernetesRuntime) Delete(ctx context.Context, resourceState *models.Resource) status.Status {
	// Validate
	if resourceState == nil {
		return status.NewErrorStatus(errors.New("resourceState is nil"))
	}

	// Get resource by attribute
	obj, resource, err := k.buildKubernetesResourceByState(resourceState)
	if err != nil {
		return status.NewErrorStatus(err)
	}

	// Delete resource
	err = resource.Delete(ctx, obj.GetName(), metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Infof("%s not found, ignore", resourceState.ResourceKey())
			return nil
		}
		return status.NewErrorStatus(err)
	}

	return nil
}

// Watch kubernetes resource by client-go
func (k *KubernetesRuntime) Watch(ctx context.Context, resourceState *models.Resource) (*models.Resource, status.Status) {
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

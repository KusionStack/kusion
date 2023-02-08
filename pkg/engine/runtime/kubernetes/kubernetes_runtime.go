package kubernetes

import (
	"context"
	"errors"

	jsonpatch "github.com/evanphx/json-patch"
	yamlv2 "gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	k8swatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/printers/k8s"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/status"
	jsonutil "kusionstack.io/kusion/pkg/util/json"
	"kusionstack.io/kusion/pkg/util/kube/config"
)

var _ runtime.Runtime = (*KubernetesRuntime)(nil)

type KubernetesRuntime struct {
	client dynamic.Interface
	mapper meta.RESTMapper
}

// NewKubernetesRuntime create a new KubernetesRuntime
func NewKubernetesRuntime() (runtime.Runtime, error) {
	client, mapper, err := getKubernetesClient()
	if err != nil {
		return nil, err
	}

	return &KubernetesRuntime{
		client: client,
		mapper: mapper,
	}, nil
}

// Apply kubernetes Resource by client-go
func (k *KubernetesRuntime) Apply(ctx context.Context, request *runtime.ApplyRequest) *runtime.ApplyResponse {
	planState := request.PlanResource
	priorState := request.PriorResource

	// Don`t consider delete case, so plan state must be not empty
	if planState == nil {
		return &runtime.ApplyResponse{Status: status.NewErrorStatus(errors.New("plan state is nil"))}
	}

	// Get kubernetes Resource interface from plan state
	planObj, resource, err := k.buildKubernetesResourceByState(planState)
	if err != nil {
		return &runtime.ApplyResponse{Status: status.NewErrorStatus(err)}
	}

	// Get live state
	response := k.Read(ctx, &runtime.ReadRequest{PlanResource: planState})
	if status.IsErr(response.Status) {
		return &runtime.ApplyResponse{Status: response.Status}
	}
	liveState := response.Resource

	// Original equals to last-applied from annotation, kusion store it in kusion_state.json
	original := ""
	if priorState != nil {
		original = jsonutil.MustMarshal2String(priorState.Attributes)
	}
	// Modified equals to input content
	modified := jsonutil.MustMarshal2String(planState.Attributes)
	// Current equals to live manifest
	current := ""
	if liveState != nil {
		current = jsonutil.MustMarshal2String(liveState.Attributes)
	}

	// Create 3-way merge patch body
	patchBody, err := jsonmergepatch.CreateThreeWayJSONMergePatch([]byte(original), []byte(modified), []byte(current))
	if err != nil {
		return &runtime.ApplyResponse{Status: status.NewErrorStatus(err)}
	}

	// Final result, dry-run to diff, otherwise to save in states
	var res *unstructured.Unstructured
	if request.DryRun {
		if liveState == nil {
			// Try ServerSideDryRun first
			createOptions := metav1.CreateOptions{
				DryRun: []string{metav1.DryRunAll},
			}
			if createdObj, err := resource.Create(ctx, planObj, createOptions); err == nil {
				res = createdObj
			} else {
				// Fall back to ClientSideDryRun
				log.Errorf("ServerSideDryRun create %s failed, fall back to ClientSideDryRun; err: %v", planState.ID, err)

				// LiveState is nil, return planObj directly
				res = planObj
			}
		} else {
			// Try ServerSideDryRun first
			patchOptions := metav1.PatchOptions{
				DryRun: []string{metav1.DryRunAll},
			}
			if patchedObj, err := resource.Patch(ctx, planObj.GetName(), types.MergePatchType, patchBody, patchOptions); err == nil {
				res = patchedObj
			} else {
				// Fall back to ClientSideDryRun
				log.Errorf("ServerSideDryRun patch %s failed, fall back to ClientSideDryRun; err: %v", planState.ID, err)

				// Merge 3-way patch
				mergedPatch, err := jsonpatch.MergePatch([]byte(current), patchBody)
				if err != nil {
					return &runtime.ApplyResponse{Status: status.NewErrorStatus(err)}
				}

				// Unmarshall and return
				res = &unstructured.Unstructured{}
				if err = res.UnmarshalJSON(mergedPatch); err != nil {
					return &runtime.ApplyResponse{Status: status.NewErrorStatus(err)}
				}
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
			return &runtime.ApplyResponse{Status: status.NewErrorStatus(err)}
		}
		// Save modified
		res = planObj
	}

	return &runtime.ApplyResponse{Resource: &models.Resource{
		ID:         planState.ResourceKey(),
		Type:       planState.Type,
		Attributes: res.Object,
		DependsOn:  planState.DependsOn,
		Extensions: planState.Extensions,
	}}
}

// Read kubernetes Resource by client-go
func (k *KubernetesRuntime) Read(ctx context.Context, request *runtime.ReadRequest) *runtime.ReadResponse {
	requestResource := request.PlanResource
	if requestResource == nil {
		requestResource = request.PriorResource
	}
	// Validate
	if requestResource == nil {
		return &runtime.ReadResponse{Status: status.NewErrorStatus(errors.New("requestResource is nil"))}
	}

	// Get resource by attribute
	obj, resource, err := k.buildKubernetesResourceByState(requestResource)
	if err != nil {
		// Ignore no match error, cause target apiVersion or kind is not installed yet
		if meta.IsNoMatchError(err) {
			log.Infof("%v, ignore", err)
			return &runtime.ReadResponse{}
		}
		return &runtime.ReadResponse{Status: status.NewErrorStatus(err)}
	}

	// Read resource
	v, err := resource.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Infof("%s not found, ignore", requestResource.ResourceKey())
			return &runtime.ReadResponse{}
		}
		return &runtime.ReadResponse{Status: status.NewErrorStatus(err)}
	}

	return &runtime.ReadResponse{Resource: &models.Resource{
		ID:         requestResource.ResourceKey(),
		Type:       requestResource.Type,
		Attributes: v.Object,
		DependsOn:  requestResource.DependsOn,
		Extensions: requestResource.Extensions,
	}}
}

// Import already exist kubernetes Resource
func (k *KubernetesRuntime) Import(ctx context.Context, request *runtime.ImportRequest) *runtime.ImportResponse {
	response := k.Read(ctx, &runtime.ReadRequest{
		PlanResource: request.PlanResource,
		Stack:        request.Stack,
	})

	if status.IsErr(response.Status) {
		return &runtime.ImportResponse{
			Resource: nil,
			Status:   response.Status,
		}
	}

	// clean up resource to make it looks like last-applied-config
	ur := &unstructured.Unstructured{Object: response.Resource.Attributes}
	lastApplied := ur.GetAnnotations()[corev1.LastAppliedConfigAnnotation]
	if len(lastApplied) != 0 {
		err := ur.UnmarshalJSON([]byte(lastApplied))
		if err != nil {
			return &runtime.ImportResponse{Status: status.NewErrorStatusWithCode(status.IllegalManifest, err)}
		}
	} else {
		unstructured.RemoveNestedField(ur.Object, "status")
		const metadata = "metadata"
		unstructured.RemoveNestedField(ur.Object, metadata, "resourceVersion")
		unstructured.RemoveNestedField(ur.Object, metadata, "creationTimestamp")
		unstructured.RemoveNestedField(ur.Object, metadata, "selfLink")
		unstructured.RemoveNestedField(ur.Object, metadata, "uid")
	}
	response.Resource.Attributes = ur.Object
	return &runtime.ImportResponse{
		Resource: response.Resource,
		Status:   nil,
	}
}

// Delete kubernetes Resource by client-go
func (k *KubernetesRuntime) Delete(ctx context.Context, request *runtime.DeleteRequest) *runtime.DeleteResponse {
	requestResource := request.Resource
	// Validate
	if requestResource == nil {
		return &runtime.DeleteResponse{Status: status.NewErrorStatus(errors.New("requestResource is nil"))}
	}

	// Get Resource by attribute
	obj, resource, err := k.buildKubernetesResourceByState(requestResource)
	if err != nil {
		return &runtime.DeleteResponse{Status: status.NewErrorStatus(err)}
	}

	// Delete Resource
	err = resource.Delete(ctx, obj.GetName(), metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Infof("%s not found, ignore", requestResource.ResourceKey())
			return &runtime.DeleteResponse{}
		}
		return &runtime.DeleteResponse{Status: status.NewErrorStatus(err)}
	}

	return &runtime.DeleteResponse{}
}

// Watch kubernetes resource by client-go
func (k *KubernetesRuntime) Watch(ctx context.Context, request *runtime.WatchRequest) *runtime.WatchResponse {
	if request == nil || request.Resource == nil {
		return &runtime.WatchResponse{Status: status.NewErrorStatus(errors.New("requestResource is nil"))}
	}

	reqObj, resource, err := k.buildKubernetesResourceByState(request.Resource)
	if err != nil {
		return &runtime.WatchResponse{Status: status.NewErrorStatus(err)}
	}

	// Root watcher
	w, err := resource.Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return &runtime.WatchResponse{Status: status.NewErrorStatus(err)}
	}
	rootCh := doWatch(ctx, w, func(watched *unstructured.Unstructured) bool {
		return watched.GetName() == reqObj.GetName()
	})

	// Collect all
	var resultChs []<-chan k8swatch.Event
	resultChs = append(resultChs, rootCh)

	if reqObj.GetKind() == k8s.Service { // Watch Endpoints or EndpointSlice
		if gvk, err := k.mapper.KindFor(schema.GroupVersionResource{
			Group:    discoveryv1.SchemeGroupVersion.Group,
			Version:  discoveryv1.SchemeGroupVersion.Version,
			Resource: k8s.EndpointSlice,
		}); gvk.Empty() || err != nil { // Watch Endpoints
			log.Errorf("k8s runtime has no kind for EndpointSlice, err: %v", err)
			namedGVK := getNamedGVK(reqObj.GroupVersionKind())
			ch, _, err := k.WatchByRelation(ctx, reqObj, namedGVK, namedBy)
			if err != nil {
				return &runtime.WatchResponse{Status: status.NewErrorStatus(err)}
			}
			resultChs = append(resultChs, ch)
		} else { // Watch EndpointSlice
			dependentGVK := getDependentGVK(reqObj.GroupVersionKind())
			ch, _, err := k.WatchByRelation(ctx, reqObj, dependentGVK, ownedBy)
			if err != nil {
				return &runtime.WatchResponse{Status: status.NewErrorStatus(err)}
			}
			resultChs = append(resultChs, ch)
		}
	} else {
		// Try to get dependent resource by owner reference
		dependentGVK := getDependentGVK(reqObj.GroupVersionKind())
		if !dependentGVK.Empty() {
			owner := reqObj
			for !dependentGVK.Empty() {
				ch, dependent, err := k.WatchByRelation(ctx, owner, dependentGVK, ownedBy)
				if err != nil {
					return &runtime.WatchResponse{Status: status.NewErrorStatus(err)}
				}
				resultChs = append(resultChs, ch)

				if dependent == nil {
					break
				}
				// Try to get deeper, max depth is 3 including root
				dependentGVK = getDependentGVK(dependent.GroupVersionKind())
				// Replace owner
				owner = dependent
			}
		}
	}

	return &runtime.WatchResponse{ResultChs: resultChs}
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

	resource, err = buildDynamicResource(k.client, k.mapper, gvk, obj.GetNamespace())
	if err != nil {
		return nil, nil, err
	}

	return obj, resource, nil
}

// buildDynamicResource get resource interface by gvk and namespace
func buildDynamicResource(
	dyn dynamic.Interface, mapper meta.RESTMapper,
	gvk *schema.GroupVersionKind, namespace string,
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
		dr = dyn.Resource(mapping.Resource).Namespace(namespace)
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

// WatchBySelector watch resources by gvk and filter by selector
func (k *KubernetesRuntime) WatchBySelector(
	ctx context.Context,
	o *unstructured.Unstructured,
	gvk schema.GroupVersionKind,
	labelStr string,
) (<-chan k8swatch.Event, error) {
	clientForResource, err := buildDynamicResource(k.client, k.mapper, &gvk, o.GetNamespace())
	if err != nil {
		return nil, err
	}

	w, err := clientForResource.Watch(ctx, metav1.ListOptions{LabelSelector: labelStr})
	if err != nil {
		return nil, err
	}

	return doWatch(ctx, w, nil), nil
}

// WatchByRelation watched resources by giving gvk if related() return true
func (k *KubernetesRuntime) WatchByRelation(
	ctx context.Context,
	cur *unstructured.Unstructured,
	gvk schema.GroupVersionKind,
	related func(watched, cur *unstructured.Unstructured) bool,
) (<-chan k8swatch.Event, *unstructured.Unstructured, error) {
	mapping, err := k.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, nil, err
	}

	clientForResource := k.client.Resource(mapping.Resource).Namespace(cur.GetNamespace())
	w, err := clientForResource.Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	var next *unstructured.Unstructured
	return doWatch(ctx, w, func(watched *unstructured.Unstructured) bool {
		ok := related(watched, cur)
		if ok {
			next = watched
		}
		return ok
	}), next, nil
}

// doWatch send watched object if check ok
func doWatch(ctx context.Context, watcher k8swatch.Interface, checker func(watched *unstructured.Unstructured) bool) <-chan k8swatch.Event {
	resultCh := make(chan k8swatch.Event)
	go func() {
		defer watcher.Stop()
		for {
			select {
			case e := <-watcher.ResultChan():
				dependent, ok := e.Object.(*unstructured.Unstructured)
				if !ok {
					break
				}
				// Check
				if checker == nil || checker != nil && checker(dependent) {
					resultCh <- e
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return resultCh
}

// Judge dependent is owned by owner
func ownedBy(dependent, owner *unstructured.Unstructured) bool {
	// Parse dependent's metadata.ownerReferences
	ownerReferences, exists, _ := unstructured.NestedSlice(dependent.Object, "metadata", "ownerReferences")
	if !exists {
		return false
	}

	// Parse owner's apiVersion, kind and metadata.name
	apiVersion := owner.GetAPIVersion()
	kind := owner.GetKind()
	name := owner.GetName()

	for _, refI := range ownerReferences {
		ref, isMap := refI.(map[string]interface{})
		if !isMap {
			continue
		}

		if ref["apiVersion"] == apiVersion && ref["kind"] == kind && ref["name"] == name {
			return true
		}
	}

	return false
}

// Service and Endpoints must be in same name and namespace
func namedBy(ep, svc *unstructured.Unstructured) bool {
	return svc.GetNamespace() == ep.GetNamespace() && svc.GetName() == ep.GetName()
}

func getDependentGVK(gvk schema.GroupVersionKind) schema.GroupVersionKind {
	switch gvk.Kind {
	// Deployment generates ReplicaSet
	case k8s.Deployment:
		return schema.GroupVersionKind{
			Group:   appsv1.SchemeGroupVersion.Group,
			Version: appsv1.SchemeGroupVersion.Version,
			Kind:    k8s.ReplicaSet,
		}
	// DaemonSet and StatefulSet generate ControllerRevision
	case k8s.DaemonSet, k8s.StatefulSet:
		return schema.GroupVersionKind{
			Group:   appsv1.SchemeGroupVersion.Group,
			Version: appsv1.SchemeGroupVersion.Version,
			Kind:    k8s.ControllerRevision,
		}
	// CronJob generates Job
	case k8s.CronJob:
		return schema.GroupVersionKind{
			Group:   batchv1.SchemeGroupVersion.Group,
			Version: batchv1.SchemeGroupVersion.Version,
			Kind:    k8s.Job,
		}
	// ReplicaSet, ReplicationController, ControllerRevision and job generate Pod
	case k8s.ReplicaSet, k8s.Job, k8s.ReplicationController, k8s.ControllerRevision:
		return schema.GroupVersionKind{
			Group:   corev1.SchemeGroupVersion.Group,
			Version: corev1.SchemeGroupVersion.Version,
			Kind:    k8s.Pod,
		}
	// Service is the owner of EndpointSlice
	case k8s.Service:
		return schema.GroupVersionKind{
			Group:   discoveryv1.SchemeGroupVersion.Group,
			Version: discoveryv1.SchemeGroupVersion.Version,
			Kind:    k8s.EndpointSlice,
		}
	default:
		return schema.GroupVersionKind{}
	}
}

func getNamedGVK(gvk schema.GroupVersionKind) schema.GroupVersionKind {
	switch gvk.Kind {
	case k8s.Service:
		return schema.GroupVersionKind{
			Group:   corev1.SchemeGroupVersion.Group,
			Version: corev1.SchemeGroupVersion.Version,
			Kind:    k8s.Endpoints,
		}
	default:
		return schema.GroupVersionKind{}
	}
}

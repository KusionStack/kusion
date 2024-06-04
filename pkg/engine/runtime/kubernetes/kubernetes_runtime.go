package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	k8swatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/printers/convertor"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes/kubeops"
	"kusionstack.io/kusion/pkg/log"
	jsonutil "kusionstack.io/kusion/pkg/util/json"
)

var _ runtime.Runtime = (*KubernetesRuntime)(nil)

type KubernetesRuntime struct {
	client dynamic.Interface
	mapper meta.RESTMapper
}

// NewKubernetesRuntime create a new KubernetesRuntime
func NewKubernetesRuntime(resource *apiv1.Resource) (runtime.Runtime, error) {
	client, mapper, err := getKubernetesClient(resource)
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
		return &runtime.ApplyResponse{Status: v1.NewErrorStatus(errors.New("plan state is nil"))}
	}

	// Get kubernetes Resource interface from plan state
	planObj, resource, err := k.buildKubernetesResourceByState(planState)
	if err != nil {
		return &runtime.ApplyResponse{Status: v1.NewErrorStatus(err)}
	}

	// Get live state
	response := k.Read(ctx, &runtime.ReadRequest{PlanResource: planState})
	if v1.IsErr(response.Status) {
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
		return &runtime.ApplyResponse{Status: v1.NewErrorStatus(err)}
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
					return &runtime.ApplyResponse{Status: v1.NewErrorStatus(err)}
				}

				// Unmarshall and return
				res = &unstructured.Unstructured{}
				if err = res.UnmarshalJSON(mergedPatch); err != nil {
					return &runtime.ApplyResponse{Status: v1.NewErrorStatus(err)}
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
			return &runtime.ApplyResponse{Status: v1.NewErrorStatus(err)}
		}
		// Save modified
		res = planObj
	}

	// Ignore the redundant fields automatically added by the K8s server for a
	// more concise and clean resource object.
	normalizeServerSideFields(res)

	// Extract the watch channel from the context.
	watchCh, _ := ctx.Value(engine.WatchChannel).(chan string)
	if !request.DryRun && watchCh != nil {
		log.Infof("Started to watch %s with the type of %s", planState.ResourceKey(), planState.Type)
		watchCh <- planState.ResourceKey()
	}

	return &runtime.ApplyResponse{Resource: &apiv1.Resource{
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
		return &runtime.ReadResponse{Status: v1.NewErrorStatus(errors.New("can not read k8s resource with empty body"))}
	}

	// Get resource by attribute
	obj, resource, err := k.buildKubernetesResourceByState(requestResource)
	if err != nil {
		// Ignore no match error, cause target apiVersion or kind is not installed yet
		if meta.IsNoMatchError(err) {
			log.Infof("%v, ignore", err)
			return &runtime.ReadResponse{}
		}
		return &runtime.ReadResponse{Status: v1.NewErrorStatus(err)}
	}

	// Read resource
	v, err := resource.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Infof("%s not found, ignore", requestResource.ResourceKey())
			return &runtime.ReadResponse{}
		}
		return &runtime.ReadResponse{Status: v1.NewErrorStatus(err)}
	}

	// Ignore the redundant fields automatically added by the K8s server for a
	// more concise and clean resource object.
	normalizeServerSideFields(v)

	return &runtime.ReadResponse{Resource: &apiv1.Resource{
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

	if v1.IsErr(response.Status) {
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
			return &runtime.ImportResponse{Status: v1.NewErrorStatusWithCode(v1.IllegalManifest, err)}
		}
	} else if convertor.Service == ur.GetKind() {
		// normalize resources
		if err := normalizeService(ur); err != nil {
			return &runtime.ImportResponse{
				Resource: nil,
				Status:   v1.NewErrorStatusWithCode(v1.IllegalManifest, err),
			}
		}
	}
	response.Resource.Attributes = ur.Object
	return &runtime.ImportResponse{
		Resource: response.Resource,
		Status:   nil,
	}
}

// normalize fields added by K8s that will cause a perpetual diff
func normalizeServerSideFields(ur *unstructured.Unstructured) {
	const metadata = "metadata"
	unstructured.RemoveNestedField(ur.Object, "status")
	unstructured.RemoveNestedField(ur.Object, metadata, "resourceVersion")
	unstructured.RemoveNestedField(ur.Object, metadata, "creationTimestamp")
	unstructured.RemoveNestedField(ur.Object, metadata, "selfLink")
	unstructured.RemoveNestedField(ur.Object, metadata, "uid")
	unstructured.RemoveNestedField(ur.Object, metadata, "generation")
	unstructured.RemoveNestedField(ur.Object, metadata, "managedFields")
}

func normalizeService(ur *unstructured.Unstructured) error {
	target := &corev1.Service{}
	if err := k8sruntime.DefaultUnstructuredConverter.FromUnstructured(ur.Object, target); err != nil {
		return err
	}
	if len(target.Spec.ClusterIPs) > 0 {
		target.Spec.ClusterIPs = nil
	}
	if len(target.Spec.ClusterIP) > 0 {
		target.Spec.ClusterIP = ""
	}
	toUnstructured, err := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(target)
	if err != nil {
		return err
	}
	ur.Object = toUnstructured
	return nil
}

// Delete kubernetes Resource by client-go
func (k *KubernetesRuntime) Delete(ctx context.Context, request *runtime.DeleteRequest) *runtime.DeleteResponse {
	requestResource := request.Resource
	// Validate
	if requestResource == nil {
		return &runtime.DeleteResponse{Status: v1.NewErrorStatus(errors.New("requestResource is nil"))}
	}

	// Get Resource by attribute
	obj, resource, err := k.buildKubernetesResourceByState(requestResource)
	if err != nil {
		return &runtime.DeleteResponse{Status: v1.NewErrorStatus(err)}
	}

	// Delete Resource
	err = resource.Delete(ctx, obj.GetName(), metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Infof("%s not found, ignore", requestResource.ResourceKey())
			return &runtime.DeleteResponse{}
		}
		return &runtime.DeleteResponse{Status: v1.NewErrorStatus(err)}
	}

	return &runtime.DeleteResponse{}
}

// Watch kubernetes resource by client-go
func (k *KubernetesRuntime) Watch(ctx context.Context, request *runtime.WatchRequest) *runtime.WatchResponse {
	if request == nil || request.Resource == nil {
		return &runtime.WatchResponse{Status: v1.NewErrorStatus(errors.New("requestResource is nil"))}
	}

	reqObj, resource, err := k.buildKubernetesResourceByState(request.Resource)
	if err != nil {
		return &runtime.WatchResponse{Status: v1.NewErrorStatus(err)}
	}

	// Root watcher
	w, err := resource.Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return &runtime.WatchResponse{Status: v1.NewErrorStatus(err)}
	}
	rootCh := doWatch(ctx, w, func(watched *unstructured.Unstructured) bool {
		return watched.GetName() == reqObj.GetName()
	})

	if rootCh == nil {
		return &runtime.WatchResponse{Status: v1.NewErrorStatus(fmt.Errorf("failed to get the root channel for watching %s",
			request.Resource.ResourceKey()))}
	}

	// Collect all
	watchers := runtime.NewWatchers()
	watchers.Insert(engine.BuildIDForKubernetes(reqObj), rootCh)

	if reqObj.GetKind() == convertor.Service { // Watch Endpoints or EndpointSlice
		if gvk, err := k.mapper.KindFor(schema.GroupVersionResource{
			Group:    discoveryv1.SchemeGroupVersion.Group,
			Version:  discoveryv1.SchemeGroupVersion.Version,
			Resource: convertor.EndpointSlice,
		}); gvk.Empty() || err != nil { // Watch Endpoints
			log.Errorf("k8s runtime has no kind for EndpointSlice, err: %v", err)
			namedGVK := getNamedGVK(reqObj.GroupVersionKind())
			ch, dependent, err := k.WatchByRelation(ctx, reqObj, namedGVK, namedBy)
			if err != nil {
				return &runtime.WatchResponse{Status: v1.NewErrorStatus(err)}
			}
			watchers.Insert(engine.BuildIDForKubernetes(dependent), ch)
		} else { // Watch EndpointSlice
			dependentGVK := getDependentGVK(reqObj.GroupVersionKind())
			ch, dependent, err := k.WatchByRelation(ctx, reqObj, dependentGVK, ownedBy)
			if err != nil {
				return &runtime.WatchResponse{Status: v1.NewErrorStatus(err)}
			}
			watchers.Insert(engine.BuildIDForKubernetes(dependent), ch)
		}
	} else {
		// Try to get dependent resource by owner reference
		dependentGVK := getDependentGVK(reqObj.GroupVersionKind())
		if !dependentGVK.Empty() {
			owner := reqObj
			for !dependentGVK.Empty() {
				ch, dependent, err := k.WatchByRelation(ctx, owner, dependentGVK, ownedBy)
				if err != nil {
					return &runtime.WatchResponse{Status: v1.NewErrorStatus(err)}
				}

				if dependent == nil {
					break
				}
				watchers.Insert(engine.BuildIDForKubernetes(dependent), ch)

				// Try to get deeper, max depth is 3 including root
				dependentGVK = getDependentGVK(dependent.GroupVersionKind())
				// Replace owner
				owner = dependent
			}
		}
	}

	return &runtime.WatchResponse{Watchers: watchers}
}

// getKubernetesClient get kubernetes client
func getKubernetesClient(resource *apiv1.Resource) (dynamic.Interface, meta.RESTMapper, error) {
	// build config
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeops.GetKubeConfig(resource))
	if err != nil {
		return nil, nil, err
	}

	// DynamicRESTMapper can discover resource types at runtime dynamically
	client, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return nil, nil, err
	}
	mapper, err := apiutil.NewDynamicRESTMapper(cfg, client)
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
func (k *KubernetesRuntime) buildKubernetesResourceByState(resourceState *apiv1.Resource) (*unstructured.Unstructured, dynamic.ResourceInterface, error) {
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

	resource, err = buildDynamicResource(k.client, k.mapper, gvk, resourceState.ID, obj.GetNamespace())
	if err != nil {
		return nil, nil, err
	}

	return obj, resource, nil
}

// buildDynamicResource get resource interface by gvk and namespace
func buildDynamicResource(
	dyn dynamic.Interface, mapper meta.RESTMapper,
	gvk *schema.GroupVersionKind, id, namespace string,
) (dynamic.ResourceInterface, error) {
	// Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	// validate whether the intent resource id matched with the GVK
	if err = validateResourceID(id, gvk); err != nil {
		return nil, err
	}

	// Obtain REST interface for the GVR
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// patch the `default` namespace for namespaced resources without explicitly
		// spcified namespace field
		keys := strings.Split(id, engine.Separator)
		if (len(keys) < 3 || keys[2] == "" || keys[2] == "default") && namespace == "" {
			namespace = "default"
		} else if len(keys) > 2 && keys[2] != namespace {
			return nil, fmt.Errorf("unmatched namespace in resource id: %s and object attribute: %s", keys[2], namespace)
		}

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
	id, labelStr string,
) (<-chan k8swatch.Event, error) {
	clientForResource, err := buildDynamicResource(k.client, k.mapper, &gvk, id, o.GetNamespace())
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

	eventCh := doWatch(ctx, w, func(watched *unstructured.Unstructured) bool {
		ok := related(watched, cur)
		if ok {
			next = watched
		}
		return ok
	})
	if eventCh == nil {
		err = fmt.Errorf("failed to get the event channel for watching related resources of %s with kind of %s",
			cur.GetName(), cur.GetKind())
	}

	return eventCh, next, err
}

// doWatch send watched object if check ok
func doWatch(ctx context.Context, watcher k8swatch.Interface, checker func(watched *unstructured.Unstructured) bool) <-chan k8swatch.Event {
	// Buffered channel, store new event
	resultCh := make(chan k8swatch.Event, 1)
	// Wait for the first watched obj
	first := true
	signal := make(chan struct{})
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
					if first {
						signal <- struct{}{}
						first = false
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Owner&Dependent check pass, return the dependent Obj
	select {
	case <-signal:
		return resultCh
	case <-ctx.Done():
		return nil
	}
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
	case convertor.Deployment:
		return schema.GroupVersionKind{
			Group:   appsv1.SchemeGroupVersion.Group,
			Version: appsv1.SchemeGroupVersion.Version,
			Kind:    convertor.ReplicaSet,
		}
	// DaemonSet and StatefulSet generate ControllerRevision
	case convertor.DaemonSet, convertor.StatefulSet:
		return schema.GroupVersionKind{
			Group:   appsv1.SchemeGroupVersion.Group,
			Version: appsv1.SchemeGroupVersion.Version,
			Kind:    convertor.ControllerRevision,
		}
	// CronJob generates Job
	case convertor.CronJob:
		return schema.GroupVersionKind{
			Group:   batchv1.SchemeGroupVersion.Group,
			Version: batchv1.SchemeGroupVersion.Version,
			Kind:    convertor.Job,
		}
	// ReplicaSet, ReplicationController, ControllerRevision and job generate Pod
	case convertor.ReplicaSet, convertor.Job, convertor.ReplicationController, convertor.ControllerRevision:
		return schema.GroupVersionKind{
			Group:   corev1.SchemeGroupVersion.Group,
			Version: corev1.SchemeGroupVersion.Version,
			Kind:    convertor.Pod,
		}
	// Service is the owner of EndpointSlice
	case convertor.Service:
		return schema.GroupVersionKind{
			Group:   discoveryv1.SchemeGroupVersion.Group,
			Version: discoveryv1.SchemeGroupVersion.Version,
			Kind:    convertor.EndpointSlice,
		}
	default:
		return schema.GroupVersionKind{}
	}
}

func getNamedGVK(gvk schema.GroupVersionKind) schema.GroupVersionKind {
	switch gvk.Kind {
	case convertor.Service:
		return schema.GroupVersionKind{
			Group:   corev1.SchemeGroupVersion.Group,
			Version: corev1.SchemeGroupVersion.Version,
			Kind:    convertor.Endpoints,
		}
	default:
		return schema.GroupVersionKind{}
	}
}

func validateResourceID(id string, gvk *schema.GroupVersionKind) error {
	keys := strings.Split(id, engine.Separator)
	if len(keys) < 2 {
		return fmt.Errorf("invalid resource id with missing required fields: %s", id)
	}

	apiVersion := keys[0]
	kind := keys[1]

	if apiVersion != gvk.GroupVersion().String() {
		return fmt.Errorf("unmatched API Version in resource id: %s and gvk: %s", apiVersion, gvk.GroupVersion().String())
	}

	if kind != gvk.Kind {
		return fmt.Errorf("unmatched Kind in resource id: %s and gvk: %s", kind, gvk.Kind)
	}

	return nil
}

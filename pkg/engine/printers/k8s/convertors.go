package k8s

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// API groups
const (
	CoreGroup      = ""
	AppsGroup      = "apps"
	DiscoveryGroup = "discovery.k8s.io"
)

// API versions
const (
	V1       = "v1"
	V1Beta1  = "v1beta1"
	V1Alpha1 = "v1alpha1"
)

// APIs in core/v1
const (
	ComponentStatus       = "ComponentStatus"
	ConfigMap             = "ConfigMap"
	Endpoints             = "Endpoints"
	Event                 = "Event"
	Namespace             = "Namespace"
	Node                  = "Node"
	PersistentVolumeClaim = "PersistentVolumeClaim"
	PersistentVolume      = "PersistentVolume"
	Pod                   = "Pod"
	PodTemplate           = "PodTemplate"
	ReplicationController = "ReplicationController"
	ResourceQuota         = "ResourceQuota"
	Secret                = "Secret"
	ServiceAccount        = "ServiceAccount"
	Service               = "Service"
)

// APIs in apps/v1
const (
	Deployment         = "Deployment"
	ReplicaSet         = "ReplicaSet"
	DaemonSet          = "DaemonSet"
	StatefulSet        = "StatefulSet"
	ControllerRevision = "ControllerRevision"
)

// APIs in discovery.k8s.io/v1
const (
	EndpointSlice = "EndpointSlice"
)

func Convert(o *unstructured.Unstructured) runtime.Object {
	switch o.GroupVersionKind().Group {
	case CoreGroup:
		return convertCore(o)
	case AppsGroup:
		return convertApps(o)
	case DiscoveryGroup:
		return convertDiscovery(o)
	default:
		return nil
	}
}

func convertCore(o *unstructured.Unstructured) runtime.Object {
	var target runtime.Object
	switch o.GetKind() {
	case ComponentStatus:
		target = &corev1.ComponentStatus{}
	case ConfigMap:
		target = &corev1.ConfigMap{}
	case Endpoints:
		target = &corev1.Endpoints{}
	case Event:
		target = &corev1.Event{}
	case Namespace:
		target = &corev1.Namespace{}
	case Node:
		target = &corev1.Node{}
	case PersistentVolumeClaim:
		target = &corev1.PersistentVolumeClaim{}
	case PersistentVolume:
		target = &corev1.PersistentVolume{}
	case Pod:
		target = &corev1.Pod{}
	case PodTemplate:
		target = &corev1.PodTemplate{}
	case ReplicationController:
		target = &corev1.ReplicationController{}
	case ResourceQuota:
		target = &corev1.ResourceQuota{}
	case Secret:
		target = &corev1.Secret{}
	case ServiceAccount:
		target = &corev1.ServiceAccount{}
	case Service:
		target = &corev1.Service{}
	default:
		return nil
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, target); err != nil {
		return nil
	}
	return target
}

func convertApps(o *unstructured.Unstructured) runtime.Object {
	var target runtime.Object
	switch o.GetKind() {
	case Deployment:
		target = &appsv1.Deployment{}
	case ReplicaSet:
		target = &appsv1.ReplicaSet{}
	case DaemonSet:
		target = &appsv1.DaemonSet{}
	case StatefulSet:
		target = &appsv1.StatefulSet{}
	case ControllerRevision:
		target = &appsv1.ControllerRevision{}
	default:
		return nil
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, target); err != nil {
		return nil
	}
	return target
}

func convertDiscovery(o *unstructured.Unstructured) runtime.Object {
	var target runtime.Object
	switch o.GetKind() {
	case EndpointSlice:
		target = &discoveryv1.EndpointSlice{}
	default:
		return nil
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, target); err != nil {
		return nil
	}
	return target
}

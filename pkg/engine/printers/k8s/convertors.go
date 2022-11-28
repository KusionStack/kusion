package k8s

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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

// APIs in batch/v1
const (
	CronJob = "CronJob"
	Job     = "Job"
)

// APIs in discovery.k8s.io/v1
const (
	EndpointSlice = "EndpointSlice"
)

func Convert(o *unstructured.Unstructured) runtime.Object {
	switch o.GroupVersionKind().GroupVersion() {
	case corev1.SchemeGroupVersion:
		return convertCoreV1(o)
	case appsv1.SchemeGroupVersion:
		return convertAppsV1(o)
	case batchv1.SchemeGroupVersion:
		return convertBatchV1(o)
	case discoveryv1.SchemeGroupVersion:
		return convertDiscoveryV1(o)
	default:
		return nil
	}
}

func convertCoreV1(o *unstructured.Unstructured) runtime.Object {
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

func convertAppsV1(o *unstructured.Unstructured) runtime.Object {
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

func convertBatchV1(o *unstructured.Unstructured) runtime.Object {
	var target runtime.Object
	switch o.GetKind() {
	case CronJob:
		target = &batchv1.CronJob{}
	case Job:
		target = &batchv1.Job{}
	default:
		return nil
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, target); err != nil {
		return nil
	}
	return target
}

func convertDiscoveryV1(o *unstructured.Unstructured) runtime.Object {
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

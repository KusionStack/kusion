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
	Namespace = "Namespace"
	Service   = "Service"
	Endpoints = "Endpoints"
)

// APIs in apps/v1
const (
	Deployment = "Deployment"
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
	case Namespace:
		target = &corev1.Namespace{}
	case Service:
		target = &corev1.Service{}
	case Endpoints:
		target = &corev1.Endpoints{}
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

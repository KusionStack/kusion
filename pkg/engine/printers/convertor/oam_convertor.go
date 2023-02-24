package convertor

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	oamv1beta1 "kusionstack.io/kusion/third_party/kubevela/kubevela/apis/v1beta1"
)

// APIs in core.oam.dev/v1beta1
const (
	Application = "Application"
)

func ToOAM(o *unstructured.Unstructured) runtime.Object {
	switch o.GroupVersionKind().GroupVersion() {
	case oamv1beta1.SchemeGroupVersion:
		return convertOamV1beta1(o)
	default:
		return nil
	}
}

func convertOamV1beta1(o *unstructured.Unstructured) runtime.Object {
	var target runtime.Object
	switch o.GetKind() {
	case Application:
		target = &oamv1beta1.Application{}
	default:
		return nil
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, target); err != nil {
		return nil
	}
	return target
}

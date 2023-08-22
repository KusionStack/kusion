package convertor

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"kusionstack.io/kube-api/apps/v1alpha1"

	"kusionstack.io/kusion/pkg/log"
)

const CollaSet = "CollaSet"

func ToKafed(u *unstructured.Unstructured) runtime.Object {
	switch u.GroupVersionKind().GroupVersion() {
	case v1alpha1.GroupVersion:
		return convertV1alpha1(u)
	default:
		return nil
	}
}

func convertV1alpha1(o *unstructured.Unstructured) runtime.Object {
	var target runtime.Object
	switch o.GetKind() {
	case CollaSet:
		target = &v1alpha1.CollaSet{}
	default:
		return nil
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, target); err != nil {
		log.Errorf("convert obj to target error. obj:%v, target:%v", o.Object, target)
		return nil
	}
	return target
}

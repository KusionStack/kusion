package printers

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"kusionstack.io/kusion/pkg/engine/printers/convertor"
)

var convertors []Convertor

type Convertor func(o *unstructured.Unstructured) runtime.Object

func init() {
	convertors = []Convertor{convertor.ToK8s, convertor.ToKafed, convertor.ToOAM}
}

func Convert(o *unstructured.Unstructured) runtime.Object {
	for _, c := range convertors {
		if target := c(o); target != nil {
			return target
		}
	}
	return nil
}

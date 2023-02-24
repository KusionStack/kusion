package printers

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"kusionstack.io/kusion/pkg/engine/printers/convertor"
)

func init() {
	registerConvertor(convertor.ToK8s)
	registerConvertor(convertor.ToOAM)
}

type Convertor func(o *unstructured.Unstructured) runtime.Object

var convertors []Convertor

func registerConvertor(c Convertor) {
	convertors = append(convertors, c)
}

func Convert(o *unstructured.Unstructured) runtime.Object {
	for _, c := range convertors {
		if target := c(o); target != nil {
			return target
		}
	}
	return nil
}

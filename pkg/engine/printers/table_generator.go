/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// ref: k8s.io/kubernetes/pkg/printers/tablegenerator.go
package printers

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

// TableGenerator - an interface for generating a message and ready flag provided a runtime.Object
type TableGenerator interface {
	GenerateTable(obj runtime.Object) (string, bool)
}

// PrintHandler - interface to handle printing provided a runtime.Object
type PrintHandler interface {
	TableHandler(printFunc interface{}) error
}

type handlerEntry struct {
	printFunc reflect.Value
}

var (
	_ TableGenerator = &HumanReadableGenerator{}
	_ PrintHandler   = &HumanReadableGenerator{}
)

// HumanReadableGenerator is an implementation of TableGenerator used to generate
// a table for a specific resource. The table is printed with a TablePrinter using
// PrintObj().
type HumanReadableGenerator struct {
	handlerMap map[reflect.Type]*handlerEntry
}

// NewTableGenerator creates a HumanReadableGenerator suitable for calling GenerateTable().
func NewTableGenerator() *HumanReadableGenerator {
	return &HumanReadableGenerator{
		handlerMap: make(map[reflect.Type]*handlerEntry),
	}
}

// With method - accepts a list of builder functions that modify HumanReadableGenerator
func (h *HumanReadableGenerator) With(fns ...func(PrintHandler)) *HumanReadableGenerator {
	for _, fn := range fns {
		fn(h)
	}
	return h
}

// GenerateTable returns a table for the provided object, using the printer registered for that type. It returns
// a table that includes all the information requested by options, but will not remove rows or columns. The
// caller is responsible for applying rules related to filtering rows or columns.
func (h *HumanReadableGenerator) GenerateTable(obj runtime.Object) (string, bool) {
	t := reflect.TypeOf(obj)
	handler, ok := h.handlerMap[t]
	if !ok {
		return "Unsupported Kind, skip", true
	}

	args := []reflect.Value{reflect.ValueOf(obj)}
	results := handler.printFunc.Call(args)

	return results[0].String(), results[1].Bool()
}

// TableHandler adds a print handler with a given set of columns to HumanReadableGenerator instance.
// See ValidateRowPrintHandlerFunc for required method signature.
func (h *HumanReadableGenerator) TableHandler(printFunc interface{}) error {
	printFuncValue := reflect.ValueOf(printFunc)
	if err := ValidateRowPrintHandlerFunc(printFuncValue); err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to register print function: %v", err))
		return err
	}
	entry := &handlerEntry{
		printFunc: printFuncValue,
	}

	objType := printFuncValue.Type().In(0)
	if _, ok := h.handlerMap[objType]; ok {
		err := fmt.Errorf("registered duplicate printer for %v", objType)
		utilruntime.HandleError(err)
		return err
	}
	h.handlerMap[objType] = entry
	return nil
}

// ValidateRowPrintHandlerFunc validates print handler signature.
// printFunc is the function that will be called to print an object.
// It must be of the following type:
//
//	func printFunc(object ObjectType) (string, bool)
//
// where ObjectType is the type of the object that will be printed, and the first
// return value is a string of key parameters of object, and second is a bool flag
// which indicates object is ready or not
func ValidateRowPrintHandlerFunc(printFunc reflect.Value) error {
	if printFunc.Kind() != reflect.Func {
		return fmt.Errorf("invalid print handler. %#v is not a function", printFunc)
	}
	funcType := printFunc.Type()
	if funcType.NumIn() != 1 || funcType.NumOut() != 2 {
		return fmt.Errorf("invalid print handler." +
			"Must accept 1 parameter and return 2 value")
	}
	if funcType.Out(0) != reflect.TypeOf((*string)(nil)).Elem() ||
		funcType.Out(1) != reflect.TypeOf((*bool)(nil)).Elem() {
		return fmt.Errorf("invalid print handler. The expected signature is: "+
			"func handler(obj %v) (string, bool)", funcType.In(0))
	}
	return nil
}

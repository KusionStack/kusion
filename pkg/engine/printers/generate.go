package printers

import (
	"k8s.io/apimachinery/pkg/runtime"

	"kusionstack.io/kusion/pkg/engine/printers/printer"
)

var tg = printer.NewTableGenerator()

func init() {
	tg.With(printer.AddK8sHandlers, printer.AddCollaSetHandlers, printer.AddOAMHandlers)
}

func Generate(obj runtime.Object) (string, bool) {
	return tg.GenerateTable(obj)
}

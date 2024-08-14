package printers

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"

	"kusionstack.io/kusion/pkg/engine/printers/printer"
	"kusionstack.io/kusion/pkg/util/kcl"
)

var tg = printer.NewTableGenerator()

func init() {
	tg.With(printer.AddK8sHandlers, printer.AddCollaSetHandlers, printer.AddOAMHandlers)
}

func Generate(obj runtime.Object) (string, bool) {
	return tg.GenerateTable(obj)
}

// PrintCustomizedHealthCheck prints customized health check result defined in the `extensions` field of the resource in the Spec.
func PrintCustomizedHealthCheck(healthPolicyCode string, resource []byte) (string, bool) {
	// Skip when health policy is empty
	if healthPolicyCode == "" {
		return "No health policy, skip", true
	}
	err := kcl.RunKCLHealthCheck(healthPolicyCode, resource)
	// Skip when health policy syntax is invalid
	if err == kcl.ErrInvalidSyntax {
		return fmt.Sprintf("health policy err: %s, skip", strings.TrimSpace(err.Error())), true
	}
	// Keep reconciling when health policy assertion failed
	if err == kcl.ErrEvaluationError {
		return "Reconciling...", false
	}
	if err != nil {
		return fmt.Sprintf("health policy err: %s", strings.TrimSpace(err.Error())), false
	}
	return "Reconciled", true
}

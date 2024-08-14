package kcl

import (
	"errors"
	"fmt"
	"strings"

	"kcl-lang.io/kcl-go/pkg/kcl"
	"kcl-lang.io/kcl-go/pkg/tools/format"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

const (
	ref                 = "res"
	EvaluationErrorStr  = "EvaluationError"
	InvalidSyntaxErrStr = "InvalidSyntax"
	UndefinedTypeErrStr = "UndefinedType"
)

var (
	ErrEvaluationError = errors.New("health check fail error")
	ErrInvalidSyntax   = errors.New("invalid syntax error")
)

// Validate if the KCL health policy has invalid syntax.
func validateKCLHealthCheck(healthPolicyCode string) error {
	_, err := kcl.Run("", kcl.WithCode(healthPolicyCode))
	if err != nil {
		if strings.Contains(err.Error(), InvalidSyntaxErrStr) {
			return ErrInvalidSyntax
		}
	}
	return nil
}

// Assemble and format the whole KCL code with yaml integration.
func assembleKCLHealthCheck(healthPolicyCode string, resource []byte) (string, error) {
	yamlStr := fmt.Sprintf(`
		import yaml
          %s = yaml.decode(%q)
	`, ref, resource)
	kclCode := yamlStr + healthPolicyCode
	kclFormatted, err := format.FormatCode(kclCode)
	if err != nil {
		return "", err
	}
	return string(kclFormatted), nil
}

// Run health check with KCL health policy during apply.
func RunKCLHealthCheck(healthPolicyCode string, resource []byte) error {
	err := validateKCLHealthCheck(healthPolicyCode)
	if err != nil {
		return err
	}
	kclCode, err := assembleKCLHealthCheck(healthPolicyCode, resource)
	if err != nil {
		return err
	}
	_, err = kcl.Run("", kcl.WithCode(kclCode))
	if err != nil && strings.Contains(err.Error(), EvaluationErrorStr) {
		if strings.Contains(err.Error(), UndefinedTypeErrStr) {
			// Distinguish the undefined error from  the evaluation error.
			errStr := strings.ReplaceAll(err.Error(), EvaluationErrorStr, "")
			return errors.New(errStr)
		}
		return ErrEvaluationError
	}
	return err
}

// Get KCL code from extensions of the resource in the Spec.
func ConvertKCLCode(healthPolicy any) (string, bool) {
	if hp, ok := healthPolicy.(map[string]any); ok {
		if code, ok := hp[v1.FieldKCLHealthCheckKCL].(string); ok {
			return code, true
		}
	}
	return "", false
}

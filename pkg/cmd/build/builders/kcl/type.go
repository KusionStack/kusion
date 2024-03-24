package kcl

import (
	kcl "kcl-lang.io/kcl-go"
)

// CompileResult is the result of a KCL compilation
type CompileResult struct {
	Documents     []kcl.KCLResult
	RawYAMLResult string
}

// NewCompileResult news a CompileResult by KCLResultList
func NewCompileResult(k *kcl.KCLResultList) *CompileResult {
	return &CompileResult{
		Documents:     k.Slice(),
		RawYAMLResult: k.GetRawYamlResult(),
	}
}

func (c *CompileResult) RawYAML() string {
	return c.RawYAMLResult
}

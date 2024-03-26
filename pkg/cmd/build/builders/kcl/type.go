package kcl

import (
	kcl "kcl-lang.io/kcl-go"
	kclpkg "kcl-lang.io/kcl-go/pkg/kcl"
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

// NewCompileResultByMapList news a CompileResult by map array
func NewCompileResultByMapList(mapList []map[string]interface{}) *CompileResult {
	documents := []kcl.KCLResult{}
	for _, mapItem := range mapList {
		documents = append(documents, kclpkg.NewResult(mapItem))
	}
	return &CompileResult{
		Documents: documents,
	}
}

func (c *CompileResult) RawYAML() string {
	return c.RawYAMLResult
}

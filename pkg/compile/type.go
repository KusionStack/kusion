package compile

import (
	kcl "kusionstack.io/kclvm-go"

	"kusionstack.io/kusion/pkg/util/yaml"
)

// The result of a KCL compilation
type CompileResult struct {
	Documents []kcl.KCLResult
}

// New a CompileResult by KCLResultList
func NewCompileResult(k *kcl.KCLResultList) *CompileResult {
	return &CompileResult{
		Documents: k.Slice(),
	}
}

// New a CompileResult by map array
func NewCompileResultByMapList(mapList []map[string]interface{}) *CompileResult {
	documents := []kcl.KCLResult{}
	for _, mapItem := range mapList {
		documents = append(documents, kcl.KCLResult(mapItem))
	}
	return &CompileResult{
		Documents: documents,
	}
}

func (c *CompileResult) YAMLString() string {
	documentList := []interface{}{}
	for _, document := range c.Documents {
		documentList = append(documentList, document)
	}
	return yaml.MergeToOneYAML(documentList...)
}

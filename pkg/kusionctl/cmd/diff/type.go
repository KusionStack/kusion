package diff

import (
	"fmt"

	yamlv3 "gopkg.in/yaml.v3"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// get ApiVersion+Kind+Namespace+Name as key
func GetGVKNNKey(r *yamlv3.Node) string {
	yamlContent, err := yamlv3.Marshal(r)
	if err != nil {
		return ""
	}

	rn, err := kyaml.Parse(string(yamlContent))
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s%s%s%s", rn.GetApiVersion(), rn.GetKind(), rn.GetNamespace(), rn.GetName())
}

type k8sDocuments []*yamlv3.Node

func (a k8sDocuments) Len() int           { return len(a) }
func (a k8sDocuments) Less(i, j int) bool { return GetGVKNNKey(a[i]) < GetGVKNNKey(a[j]) }
func (a k8sDocuments) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

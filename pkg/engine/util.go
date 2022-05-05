package engine

import (
	"kusionstack.io/KCLVM/kclvm-go/api/kcl"

	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/log"
	kyaml "kusionstack.io/kusion/pkg/util/yaml"
)

const (
	DefaultResourceStateMode = "managed"
	Separator                = ":"
)

func NewRequestResourceForKubernetes(r kcl.KCLResult) (*states.ResourceState, string, error) {
	// Get kubernetes manifestations, such as kind, metadata.name, metadata.namespace etc
	docs, err := kyaml.YAML2Documents(r.YAMLString())
	if err != nil {
		return nil, "", err
	}

	if len(docs) > 1 {
		log.Warn("document size is greater than 1")
	}

	doc := docs[0]

	// Parse kubernetes resource
	apiVersion, err := kyaml.GetByPathString(doc, "$.apiVersion")
	if err != nil {
		return nil, "", err
	}

	kind, err := kyaml.GetByPathString(doc, "$.kind")
	if err != nil {
		return nil, "", err
	}

	metadataName, err := kyaml.GetByPathString(doc, "$.metadata.name")
	if err != nil {
		return nil, "", err
	}

	metadataNamespace, _ := kyaml.GetByPathString(doc, "$.metadata.namespace")

	// Build request resource for kubernetes
	return &states.ResourceState{
		ID:         BuildIDForKubernetes(apiVersion, kind, metadataNamespace, metadataName),
		Mode:       DefaultResourceStateMode,
		Attributes: r,
		DependsOn:  nil,
	}, kind, nil
}

func BuildIDForKubernetes(apiVersion, kind, namespace, name string) string {
	key := apiVersion + Separator + kind + Separator
	if namespace != "" {
		key += namespace + Separator
	}
	return key + name
}

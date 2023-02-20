package kcl

import (
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util/io"
	kyaml "kusionstack.io/kusion/pkg/util/yaml"
)

const (
	KusionKclPathEnv = "KUSION_KCL_PATH"
	ID               = "id"
	Type             = "type"
	Attributes       = "attributes"
)

var kclAppPath = getKclPath()

func getKclPath() string {
	// 1. try ${KusionKclPathEnv}
	if kclPath := os.Getenv(KusionKclPathEnv); kclPath != "" {
		return kclPath
	}

	// 2.1 try ${appPath}/kclvm/bin/kcl
	// 2.2 try ${appPath}/../kclvm/bin/kcl
	// 2.3 try ${PWD}/kclvm/bin/kcl

	var kclPathList []string
	if appPath, _ := os.Executable(); appPath != "" {
		kclPathList = append(kclPathList,
			filepath.Join(filepath.Dir(appPath), "kclvm", "bin", "kcl"),
			filepath.Join(filepath.Dir(filepath.Dir(appPath)), "kclvm", "bin", "kcl"),
		)
	}
	if workDir, _ := os.Getwd(); workDir != "" {
		kclPathList = append(kclPathList,
			filepath.Join(workDir, "kclvm", "bin", "kcl"),
		)
	}
	for _, kclPath := range kclPathList {
		if ok, _ := io.IsFile(kclPath); ok {
			return kclPath
		}
	}

	// 3. try ${PATH}/kcl

	if kclPath, _ := exec.LookPath("kcl"); kclPath != "" {
		return kclPath
	}

	return "kcl"
}

func k8sResource2ResourceMap(resource map[string]interface{}) (map[string]interface{}, error) {
	// Get kubernetes manifestations, such as kind, metadata.name, metadata.namespace etc
	resourceYaml, _ := yaml.Marshal(resource)
	docs, err := kyaml.YAML2Documents(string(resourceYaml))
	if err != nil {
		return nil, err
	}

	if len(docs) > 1 {
		log.Warn("document size is greater than 1")
	}

	doc := docs[0]

	// Parse kubernetes resource
	apiVersion, err := kyaml.GetByPathString(doc, "$.apiVersion")
	if err != nil {
		return nil, err
	}

	kind, err := kyaml.GetByPathString(doc, "$.kind")
	if err != nil {
		return nil, err
	}

	metadataName, err := kyaml.GetByPathString(doc, "$.metadata.name")
	if err != nil {
		return nil, err
	}

	metadataNamespace, _ := kyaml.GetByPathString(doc, "$.metadata.namespace")

	return map[string]interface{}{
		ID:         engine.BuildID(apiVersion, kind, metadataNamespace, metadataName),
		Type:       generator.Kubernetes,
		Attributes: resource,
	}, nil
}

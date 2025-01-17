package api

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/liu-hm19/pterm"
	yamlv3 "gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/api/generate/generator"
	"kusionstack.io/kusion/pkg/engine/api/generate/run"

	// "kusionstack.io/kusion/pkg/engine/api/builders/kcl"

	"kusionstack.io/kusion/pkg/util/pretty"
)

const JSONOutput = "json"

// TODO: This switch logic may still be needed for KCL Builder
// func Intent(o *builders.Options, p *v1.Project, s *v1.Stack, ws *v1.Workspace) (*v1.Intent, error) {
// 	// Choose the generator
// 	var builder builders.Builder
// 	pg := p.Generator

// 	// default AppsConfigBuilder
// 	var bt v1.BuilderType
// 	if pg == nil {
// 		bt = v1.AppConfigurationBuilder
// 	} else {
// 		bt = pg.Type
// 	}

// 	// we can add more generators here
// 	switch bt {
// 	case v1.KCLBuilder:
// 		builder = &kcl.Builder{}
// 	case v1.AppConfigurationBuilder:
// 		appConfigs, err := buildAppConfigs(o, s)
// 		if err != nil {
// 			return nil, err
// 		}
// 		builder = &builders.AppsConfigBuilder{
// 			Apps:      appConfigs,
// 			Workspace: ws,
// 		}
// 	default:
// 		return nil, fmt.Errorf("unknow generator type:%s", bt)
// 	}

// 	i, err := builder.Build(o, p, s)
// 	if err != nil {
// 		return nil, errors.New(stripansi.Strip(err.Error()))
// 	}
// 	return i, nil
// }

// GenerateSpecWithSpinner calls generator to generate versioned Spec. Add a method wrapper for testing purposes.
func GenerateSpecWithSpinner(project *v1.Project, stack *v1.Stack, workspace *v1.Workspace, noStyle bool) (*v1.Spec, error) {
	// Construct generator instance
	defaultGenerator := &generator.DefaultGenerator{
		Project:   project,
		Stack:     stack,
		Workspace: workspace,
		Runner:    &run.KPMRunner{},
	}

	var sp *pterm.SpinnerPrinter
	if noStyle {
		fmt.Printf("Generating Spec in the Stack %s...\n", stack.Name)
	} else {
		sp = &pretty.SpinnerT
		sp, _ = sp.Start(fmt.Sprintf("Generating Spec in the Stack %s...", stack.Name))
	}

	// style means color and prompt here. Currently, sp will be nil only when o.NoStyle is true
	style := !noStyle && sp != nil

	versionedSpec, err := defaultGenerator.Generate(stack.Path, nil)
	if err != nil {
		if style {
			sp.Fail()
			return nil, err
		} else {
			return nil, err
		}
	}

	err = ValidateSpec(versionedSpec)
	if err != nil {
		return nil, err
	}

	// success
	if style {
		sp.Success()
	} else {
		fmt.Println()
	}

	return versionedSpec, nil
}

func SpecFromFile(filePath string) (*v1.Spec, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// TODO: here we use decoder in yaml.v3 to parse resources because it converts
	// map into map[string]interface{} by default which is inconsistent with yaml.v2.
	// The use of yaml.v2 and yaml.v3 should be unified in the future.
	decoder := yamlv3.NewDecoder(bytes.NewBuffer(b))
	decoder.KnownFields(true)
	i := &v1.Spec{}
	if err = decoder.Decode(i); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to parse the intent file, please check if the file content is valid")
	}
	return i, nil
}

func ValidateSpec(spec *v1.Spec) error {
	if spec == nil {
		return fmt.Errorf("spec is nil")
	}
	if len(spec.Resources) == 0 {
		return fmt.Errorf("spec has no resources")
	}
	err := validateSpecResources(spec.Resources)
	if err != nil {
		return err
	}
	return nil
}

func validateSpecResources(resources []v1.Resource) error {
	// Check for duplicate resource ids
	resourceExists := make(map[string]bool)
	for idx, resource := range resources {
		// Check for empty resource ID
		if resource.ID == "" {
			return fmt.Errorf("resource ID is empty for resource %v", idx)
		}
		// Check whether resource ID already exists
		if _, ok := resourceExists[resource.ID]; ok {
			return fmt.Errorf("duplicate resource ID %s", resource.ID)
		}
		resourceExists[resource.ID] = true
		// Check whether resource ID is valid
		err := validateResourceID(resource)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateResourceID(resource v1.Resource) error {
	switch {
	case resource.Type == v1.Kubernetes:
		return validateKubernetesResource(resource)
	case resource.Type == v1.Terraform:
		return validateTerraformResource(resource)
	default:
		return fmt.Errorf("invalid resource type: %s", resource.Type)
	}
}

func validateKubernetesResource(resource v1.Resource) error {
	idParts := strings.Split(resource.ID, engine.Separator)
	if len(idParts) < 3 || len(idParts) > 4 {
		return fmt.Errorf("invalid resource id with missing required fields: %s", resource.ID)
	}
	apiVersion := idParts[0]
	kind := idParts[1]
	if attributeAPIVersion, ok := resource.Attributes["apiVersion"]; ok {
		if attributeAPIVersion != apiVersion {
			return fmt.Errorf("unmatched API Version in resource id: %s and attribute: %s", apiVersion, attributeAPIVersion)
		}
	}
	if attributeKind, ok := resource.Attributes["kind"]; ok {
		if attributeKind != kind {
			return fmt.Errorf("unmatched Kind in resource id: %s and attribute: %s", kind, attributeKind)
		}
	}
	return nil
}

func validateTerraformResource(resource v1.Resource) error {
	idParts := strings.Split(resource.ID, engine.Separator)
	if len(idParts) != 4 {
		return fmt.Errorf("invalid resource id with missing required fields: %s", resource.ID)
	}
	var providerNamespace, providerName string
	// TODO: add provider namespace and name validation
	// if _, ok := resource.Extensions["required_provider"]; ok {
	// 	// TODO: implement when required_provider is supported in module-framework
	// }
	if providerExtension, ok := resource.Extensions["provider"].(string); ok {
		srcAttrs := strings.Split(providerExtension, "/")
		if len(srcAttrs) == 4 {
			providerNamespace = srcAttrs[1]
			providerName = srcAttrs[2]
		} else if len(srcAttrs) == 3 {
			providerNamespace = srcAttrs[0]
			providerName = srcAttrs[1]
		} else {
			return fmt.Errorf("invalid terraform provider source: %s", providerExtension)
		}
	} else {
		return fmt.Errorf("missing provider extension in terraform resource: %s", resource.ID)
	}
	if providerNamespace != idParts[0] || providerName != idParts[1] {
		return fmt.Errorf("unmatched provider in resource id: %s and provider extension: %s", idParts[0]+"/"+idParts[1], providerNamespace+"/"+providerName)
	}
	return nil
}

// func buildAppConfigs(o *builders.Options, stack *v1.Stack) (map[string]v1.AppConfiguration, error) {
// 	o.Arguments[kcl.IncludeSchemaTypePath] = "true"
// 	compileResult, err := kcl.Run(o, stack)
// 	if err != nil {
// 		return nil, err
// 	}

// 	documents := compileResult.Documents
// 	if len(documents) == 0 {
// 		return nil, fmt.Errorf("no AppConfiguration is found in the compile result")
// 	}

// 	out := documents[0].YAMLString()

// 	log.Debugf("unmarshal %s to app configs", out)
// 	appConfigs := map[string]v1.AppConfiguration{}

// 	// Note: we use the type of MapSlice in yaml.v2 to maintain the order of container
// 	// environment variables, thus we unmarshal appConfigs with yaml.v2 rather than yaml.v3.
// 	err = yaml.Unmarshal([]byte(out), appConfigs)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return appConfigs, nil
// }

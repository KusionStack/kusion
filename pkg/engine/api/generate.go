package api

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/liu-hm19/pterm"
	yamlv3 "gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
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

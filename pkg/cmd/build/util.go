package build

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/acarl005/stripansi"
	"github.com/pterm/pterm"
	"gopkg.in/yaml.v2"
	yamlv3 "gopkg.in/yaml.v3"

	"kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/cmd/build/builders"
	"kusionstack.io/kusion/pkg/cmd/build/builders/kcl"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/util/pretty"
	"kusionstack.io/kusion/pkg/workspace"
)

func IntentWithSpinner(o *builders.Options, project *v1.Project, stack *v1.Stack) (*v1.Intent, error) {
	var sp *pterm.SpinnerPrinter
	if o.NoStyle {
		fmt.Printf("Generating Intent in the Stack %s...\n", stack.Name)
	} else {
		sp = &pretty.SpinnerT
		sp, _ = sp.Start(fmt.Sprintf("Generating Intent in the Stack %s...", stack.Name))
	}

	// style means color and prompt here. Currently, sp will be nil only when o.NoStyle is true
	style := !o.NoStyle && sp != nil

	i, err := Intent(o, project, stack)
	// failed
	if err != nil {
		if style {
			sp.Fail()
			return nil, err
		} else {
			// TODO: we will replace this implementation with KCL no-style flag when it is supported
			return nil, errors.New(stripansi.Strip(err.Error()))
		}
	}

	// success
	if style {
		sp.Success()
	} else {
		fmt.Println()
	}
	return i, nil
}

func Intent(o *builders.Options, p *v1.Project, s *v1.Stack) (*v1.Intent, error) {
	// Choose the generator
	var builder builders.Builder
	pg := p.Generator

	// default AppsConfigBuilder
	var bt v1.BuilderType
	if pg == nil {
		bt = v1.AppConfigurationBuilder
	} else {
		bt = pg.Type
	}

	// we can add more generators here
	switch bt {
	case v1.KCLBuilder:
		builder = &kcl.Builder{}
	case v1.AppConfigurationBuilder:
		appConfigs, err := buildAppConfigs(o, s)
		if err != nil {
			return nil, err
		}
		ws, err := workspace.GetWorkspaceByDefaultOperator(s.Name)
		if err != nil {
			return nil, err
		}
		builder = &builders.AppsConfigBuilder{
			Apps:      appConfigs,
			Workspace: ws,
		}
	default:
		return nil, fmt.Errorf("unknow generator type:%s", bt)
	}

	i, err := builder.Build(o, p, s)
	if err != nil {
		return nil, errors.New(stripansi.Strip(err.Error()))
	}
	return i, nil
}

func buildAppConfigs(o *builders.Options, stack *v1.Stack) (map[string]inputs.AppConfiguration, error) {
	o.Arguments[kcl.IncludeSchemaTypePath] = "true"
	compileResult, err := kcl.Run(o, stack)
	if err != nil {
		return nil, err
	}

	documents := compileResult.Documents
	if len(documents) == 0 {
		return nil, fmt.Errorf("no AppConfiguration is found in the compile result")
	}

	out := documents[0].YAMLString()

	log.Debugf("unmarshal %s to app configs", out)
	appConfigs := map[string]inputs.AppConfiguration{}

	// Note: we use the type of MapSlice in yaml.v2 to maintain the order of container
	// environment variables, thus we unmarshal appConfigs with yaml.v2 rather than yaml.v3.
	err = yaml.Unmarshal([]byte(out), appConfigs)
	if err != nil {
		return nil, err
	}

	return appConfigs, nil
}

func IntentFromFile(filePath string) (*v1.Intent, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// TODO: here we use decoder in yaml.v3 to parse resources because it converts
	// map into map[string]interface{} by default which is inconsistent with yaml.v2.
	// The use of yaml.v2 and yaml.v3 should be unified in the future.
	decoder := yamlv3.NewDecoder(bytes.NewBuffer(b))
	decoder.KnownFields(true)
	i := &v1.Intent{}
	if err = decoder.Decode(i); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to parse the intent file, please check if the file content is valid")
	}
	return i, nil
}

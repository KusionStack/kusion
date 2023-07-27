package spec

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/acarl005/stripansi"
	"github.com/pterm/pterm"

	"gopkg.in/yaml.v3"

	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/generator/kcl"
	models2 "kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/util/pretty"
)

func GenerateSpecWithSpinner(o *generator.Options, project *projectstack.Project, stack *projectstack.Stack) (*models2.Spec, error) {
	var sp *pterm.SpinnerPrinter
	if o.NoStyle {
		fmt.Printf("Generating Spec in the Stack %s...\n", stack.Name)
	} else {
		sp = &pretty.SpinnerT
		sp, _ = sp.Start(fmt.Sprintf("Generating Spec in the Stack %s...", stack.Name))
	}

	// style means color and prompt here. Currently, sp will be nil only when o.NoStyle is true
	style := !o.NoStyle && sp != nil

	spec, err := GenerateSpec(o, project, stack)
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
	return spec, nil
}

func GenerateSpec(o *generator.Options, project *projectstack.Project, stack *projectstack.Stack) (*models2.Spec, error) {
	// Choose the generator
	var g generator.Generator
	pg := project.Generator

	// default Generator
	if pg == nil {
		g = &kcl.Generator{}
	} else {
		gt := pg.Type
		// we can add more generators here
		switch gt {
		case projectstack.KCLGenerator:
			g = &kcl.Generator{}
		default:
			return nil, fmt.Errorf("unknow generator type:%s", gt)
		}
	}

	spec, err := g.GenerateSpec(o, stack)
	if err != nil {
		return nil, errors.New(stripansi.Strip(err.Error()))
	}
	return spec, nil
}

func GenerateSpecFromFile(filePath string) (*models2.Spec, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	decoder := yaml.NewDecoder(bytes.NewBuffer(b))
	decoder.KnownFields(true)
	var resources models2.Resources
	if err = decoder.Decode(&resources); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to parse the spec file, please check if the file content is valid")
	}
	return &models2.Spec{Resources: resources}, nil
}

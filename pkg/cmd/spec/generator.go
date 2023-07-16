package spec

import (
	"errors"
	"fmt"

	"github.com/acarl005/stripansi"
	"github.com/pterm/pterm"

	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/generator/kcl"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/util/pretty"
)

func GenerateSpecWithSpinner(o *generator.Options, project *projectstack.Project, stack *projectstack.Stack) (*models.Spec, error) {
	var sp *pterm.SpinnerPrinter
	if o.NoStyle {
		fmt.Printf("Generating Spec in the Stack %s...\n", stack.Name)
	} else {
		sp = &pretty.SpinnerT
		sp, _ = sp.Start(fmt.Sprintf("Generating Spec in the Stack %s...", stack.Name))
	}

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
		if !o.NoStyle && sp != nil {
			sp.Fail()
		}

		// TODO: we will replace this implementation with KCL no-style flag
		// when it is supported
		if o.NoStyle {
			return nil, errors.New(stripansi.Strip(err.Error()))
		}

		return nil, err
	}

	if !o.NoStyle && sp != nil {
		sp.Success()
	}
	if !o.NoStyle {
		fmt.Println()
	}

	return spec, nil
}

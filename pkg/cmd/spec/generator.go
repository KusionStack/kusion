package spec

import (
	"fmt"

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
		if sp != nil {
			sp.Fail()
		}
		return nil, err
	}

	if sp != nil {
		sp.Success()
	}
	fmt.Println()

	return spec, nil
}

package operation

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform/dag"

	"kusionstack.io/kusion/pkg/engine/states"
)

func TestDeleteResourceParser_Parse(t *testing.T) {
	const VPC = "vpc"
	const VSwitch = "vswitch"
	const VSecutiry = "vsecurity"
	const Instance = "instance"
	resources := []states.ResourceState{
		{
			ID:   VPC,
			Mode: states.Managed,
			Attributes: map[string]interface{}{
				"c": "d",
			},
			DependsOn: nil,
		},
		{
			ID:   VSwitch,
			Mode: states.Managed,
			Attributes: map[string]interface{}{
				"a": "c",
			},
			DependsOn: []string{VPC},
		},
		{
			ID:   VSecutiry,
			Mode: states.Managed,
			Attributes: map[string]interface{}{
				"a": "b",
			},
			DependsOn: []string{VSwitch},
		},
		{
			ID:   Instance,
			Mode: states.Managed,
			Attributes: map[string]interface{}{
				"a": "b",
			},
			DependsOn: []string{VSecutiry, VSwitch},
		},
	}

	graph := &dag.AcyclicGraph{}
	graph.Add(&RootNode{})

	deleteResourceParser := &DeleteResourceParser{
		resources: resources,
	}

	_ = deleteResourceParser.Parse(graph)
	actual := strings.TrimSpace(graph.String())
	expected := strings.TrimSpace(testGraphTransReductionMultiple)

	if actual != expected {
		t.Errorf("wrong result\ngot:\n%s\n\nwant:\n%s", actual, expected)
	}
}

const testGraphTransReductionMultiple = `
instance
  vsecurity
root
  instance
vpc
vsecurity
  vswitch
vswitch
  vpc
`

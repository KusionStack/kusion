package operation

import (
	"kusionstack.io/kusion/pkg/engine/models"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/dag"
)

func TestDeleteResourceParser_Parse(t *testing.T) {
	const VPC = "vpc"
	const VSwitch = "vswitch"
	const VSecutiry = "vsecurity"
	const Instance = "instance"
	resources := []models.Resource{
		{
			ID: VPC,

			Attributes: map[string]interface{}{
				"c": "d",
			},
			DependsOn: nil,
		},
		{
			ID: VSwitch,

			Attributes: map[string]interface{}{
				"a": "c",
			},
			DependsOn: []string{VPC},
		},
		{
			ID: VSecutiry,

			Attributes: map[string]interface{}{
				"a": "b",
			},
			DependsOn: []string{VSwitch},
		},
		{
			ID: Instance,

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

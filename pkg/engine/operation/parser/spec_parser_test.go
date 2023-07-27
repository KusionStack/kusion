package parser

import (
	"strings"
	"testing"

	"kusionstack.io/kusion/pkg/engine/operation/graph"
	"kusionstack.io/kusion/third_party/terraform/dag"
)

func TestSpecParser_Parse(t *testing.T) {
	const Jack = "jack"
	const Pony = "pony"
	const Eric = "eric"
	mf := &models.Spec{Resources: []models.Resource{
		{
			ID: Pony,

			Attributes: map[string]interface{}{
				"c": "d",
			},
			DependsOn: []string{Jack},
		},
		{
			ID: Eric,

			Attributes: map[string]interface{}{
				"a": graph.ImplicitRefPrefix + "jack.a",
			},
			DependsOn: []string{Pony},
		},
		{
			ID: Jack,

			Attributes: map[string]interface{}{
				"a": "b",
			},
			DependsOn: nil,
		},
	}}

	ag := &dag.AcyclicGraph{}
	ag.Add(&graph.RootNode{})

	spec := &SpecParser{
		spec: mf,
	}

	_ = spec.Parse(ag)
	actual := strings.TrimSpace(ag.String())
	expected := strings.TrimSpace(testGraphTransReductionMultipleRootsStr)

	if actual != expected {
		t.Errorf("wrong result\ngot:\n%s\n\nwant:\n%s", actual, expected)
	}
}

const testGraphTransReductionMultipleRootsStr = `
eric
jack
  pony
pony
  eric
root
  jack
`

package operation

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform/dag"
	"kusionstack.io/kusion/pkg/engine/models"
)

func TestManifestParser_Parse(t *testing.T) {
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
				"a": ImplicitRefPrefix + "jack.a",
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

	graph := &dag.AcyclicGraph{}
	graph.Add(&RootNode{})

	manifest := &ManifestParser{
		manifest: mf,
	}

	_ = manifest.Parse(graph)
	actual := strings.TrimSpace(graph.String())
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

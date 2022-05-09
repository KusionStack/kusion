package operation

import (
	"strings"
	"testing"

	"kusionstack.io/kusion/pkg/engine/manifest"
	"kusionstack.io/kusion/pkg/engine/states"

	"github.com/hashicorp/terraform/dag"
)

func TestManifestParser_Parse(t *testing.T) {
	const Jack = "jack"
	const Pony = "pony"
	const Eric = "eric"
	mf := &manifest.Manifest{Resources: []states.ResourceState{
		{
			ID:   Pony,
			Mode: states.Managed,
			Attributes: map[string]interface{}{
				"c": "d",
			},
			DependsOn: []string{Jack},
		},
		{
			ID:   Eric,
			Mode: states.Managed,
			Attributes: map[string]interface{}{
				"a": ImplicitRefPrefix + "jack.a",
			},
			DependsOn: []string{Pony},
		},
		{
			ID:   Jack,
			Mode: states.Managed,
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

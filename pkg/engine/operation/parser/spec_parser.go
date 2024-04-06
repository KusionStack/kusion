package parser

import (
	"fmt"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation/graph"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/util"
	"kusionstack.io/kusion/pkg/util/json"
	"kusionstack.io/kusion/third_party/terraform/dag"
)

type IntentParser struct {
	intent *apiv1.Spec
}

func NewIntentParser(i *apiv1.Spec) *IntentParser {
	return &IntentParser{intent: i}
}

var _ Parser = (*IntentParser)(nil)

func (m *IntentParser) Parse(g *dag.AcyclicGraph) (s v1.Status) {
	util.CheckNotNil(g, "dag is nil")
	i := m.intent
	util.CheckNotNil(i, "models is nil")
	if i.Resources == nil {
		sprintf := fmt.Sprintf("no resources in models:%s", json.Marshal2String(i))
		return v1.NewBaseStatus(v1.Warning, v1.NotFound, sprintf)
	}

	root, err := g.Root()
	util.CheckNotError(err, "get dag root error")
	util.CheckNotNil(root, fmt.Sprintf("No root in this DAG:%s", json.Marshal2String(g)))
	resourceIndex := i.Resources.Index()
	for key, resource := range resourceIndex {
		rn, s := graph.NewResourceNode(key, resourceIndex[key], models.Update)
		if v1.IsErr(s) {
			return s
		}

		// add this resource to dag at first time
		if !g.HasVertex(rn) {
			g.Add(rn)
			g.Connect(dag.BasicEdge(root, rn))
		} else {
			// always get the latest vertex in this g otherwise you will get subtle mistake in walking this g
			rn = GetVertex(g, rn).(*graph.ResourceNode)
			g.Connect(dag.BasicEdge(root, rn))
		}

		// compute implicit and explicate dependencies
		refNodeKeys, s := updateDependencies(resource)
		if v1.IsErr(s) {
			return s
		}

		// linkRefNodes
		s = LinkRefNodes(g, refNodeKeys, resourceIndex, rn, models.Update, nil)
		if v1.IsErr(s) {
			return s
		}
	}

	if err = g.Validate(); err != nil {
		return v1.NewErrorStatusWithMsg(v1.IllegalManifest, "Found circle dependency in models:"+err.Error())
	}
	g.TransitiveReduction()
	return s
}

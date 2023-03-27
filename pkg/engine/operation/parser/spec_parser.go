package parser

import (
	"fmt"

	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/operation/graph"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/status"
	"kusionstack.io/kusion/pkg/util"
	"kusionstack.io/kusion/pkg/util/json"
	"kusionstack.io/kusion/third_party/terraform/dag"
)

type SpecParser struct {
	spec *models.Spec
}

func NewSpecParser(spec *models.Spec) *SpecParser {
	return &SpecParser{spec: spec}
}

var _ Parser = (*SpecParser)(nil)

func (m *SpecParser) Parse(g *dag.AcyclicGraph) (s status.Status) {
	util.CheckNotNil(g, "dag is nil")
	sp := m.spec
	util.CheckNotNil(sp, "models is nil")
	if sp.Resources == nil {
		sprintf := fmt.Sprintf("no resources in models:%s", json.Marshal2String(sp))
		return status.NewBaseStatus(status.Warning, status.NotFound, sprintf)
	}

	root, err := g.Root()
	util.CheckNotError(err, "get dag root error")
	util.CheckNotNil(root, fmt.Sprintf("No root in this DAG:%s", json.Marshal2String(g)))
	resourceIndex := sp.Resources.Index()
	for key, resource := range resourceIndex {
		rn, s := graph.NewResourceNode(key, resourceIndex[key], opsmodels.Update)
		if status.IsErr(s) {
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
		if status.IsErr(s) {
			return s
		}

		// linkRefNodes
		s = LinkRefNodes(g, refNodeKeys, resourceIndex, rn, opsmodels.Update, nil)
		if status.IsErr(s) {
			return s
		}
	}

	if err = g.Validate(); err != nil {
		return status.NewErrorStatusWithMsg(status.IllegalManifest, "Found circle dependency in models:"+err.Error())
	}
	g.TransitiveReduction()
	return s
}

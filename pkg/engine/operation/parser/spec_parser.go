package parser

import (
	"fmt"
	"reflect"

	"kusionstack.io/kusion/pkg/engine/operation/graph"
	"kusionstack.io/kusion/third_party/terraform/dag"

	"kusionstack.io/kusion/pkg/engine/operation/types"

	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/status"
	"kusionstack.io/kusion/pkg/util"
	"kusionstack.io/kusion/pkg/util/json"
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
	for key, resourceState := range resourceIndex {
		rn, s := graph.NewResourceNode(key, resourceIndex[key], types.Update)
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
		// handle explicate dependency
		refNodeKeys := resourceState.DependsOn

		// handle implicit dependency
		v := reflect.ValueOf(resourceState.Attributes)
		implicitRefKeys, _, s := graph.ParseImplicitRef(v, nil, func(map[string]*models.Resource, string) (reflect.Value, status.Status) {
			return v, nil
		})
		if status.IsErr(s) {
			return s
		}
		refNodeKeys = append(refNodeKeys, implicitRefKeys...)

		// Deduplicate
		refNodeKeys = Deduplicate(refNodeKeys)

		// linkRefNodes
		s = LinkRefNodes(g, refNodeKeys, resourceIndex, rn, types.Update, nil)
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

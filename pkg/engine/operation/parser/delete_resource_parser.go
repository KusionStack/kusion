package parser

import (
	"fmt"

	"kusionstack.io/kusion/pkg/engine/operation/graph"
	dag2 "kusionstack.io/kusion/third_party/terraform/dag"

	"kusionstack.io/kusion/pkg/engine/operation/types"

	"kusionstack.io/kusion/pkg/engine/models"

	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/status"
	"kusionstack.io/kusion/pkg/util"
	"kusionstack.io/kusion/pkg/util/json"
)

type DeleteResourceParser struct {
	resources models.Resources
}

func NewDeleteResourceParser(resources models.Resources) *DeleteResourceParser {
	return &DeleteResourceParser{resources: resources}
}

func (d *DeleteResourceParser) Parse(g *dag2.AcyclicGraph) (s status.Status) {
	util.CheckNotNil(g, "graph is nil")
	if len(g.Vertices()) == 0 {
		log.Infof("no vertices in dag when parsing deleted resources. dag:%s", json.Marshal2String(g))
		return status.NewErrorStatusWithMsg(status.InvalidArgument, "no vertices in dag when parsing deleted resources")
	}

	manifestGraphMap := make(map[string]interface{})
	for _, v := range g.Vertices() {
		if rn, ok := v.(*graph.ResourceNode); ok {
			id := rn.Hashcode().(string)
			manifestGraphMap[id] = v
		}
	}

	// diff resources to delete
	resourceIndex := d.resources.Index()
	root, err := g.Root()
	util.CheckNotError(err, "get dag root error")

	priorDependsOn := make(map[string][]string)
	for key, v := range resourceIndex {
		for _, dp := range v.DependsOn {
			priorDependsOn[dp] = append(priorDependsOn[dp], key)
		}
	}

	for key, resource := range resourceIndex {
		rn, s := graph.NewResourceNode(key, resourceIndex[key], types.Delete)
		if status.IsErr(s) {
			return s
		}
		rnID := rn.Hashcode().(string)

		if !g.HasVertex(rn) && manifestGraphMap[rnID] == nil {
			log.Infof("resource:%v not found in models. Mark as delete node", key)
			// we cannot delete this node if any node dependsOn this node
			for _, v := range priorDependsOn[rnID] {
				if manifestGraphMap[v] != nil {
					msg := fmt.Sprintf("%s dependson %s, cannot delete resource %s", v, rnID, rnID)
					return status.NewErrorStatusWithMsg(status.Internal, msg)
				}
			}
			g.Add(rn)
			g.Connect(dag2.BasicEdge(root, rn))
		}

		// always get the latest vertex in the g.
		rn = GetVertex(g, rn).(*graph.ResourceNode)
		s = LinkRefNodes(g, resource.DependsOn, resourceIndex, rn, types.Delete, manifestGraphMap)
		if status.IsErr(s) {
			return s
		}
	}
	if err = g.Validate(); err != nil {
		return status.NewErrorStatusWithMsg(status.IllegalManifest, "Found circle dependency in models."+err.Error())
	}
	g.TransitiveReduction()
	return s
}

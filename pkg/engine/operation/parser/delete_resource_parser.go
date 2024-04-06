package parser

import (
	"fmt"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation/graph"
	models "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util"
	"kusionstack.io/kusion/pkg/util/json"
	"kusionstack.io/kusion/third_party/terraform/dag"
)

type DeleteResourceParser struct {
	resources apiv1.Resources
}

func NewDeleteResourceParser(resources apiv1.Resources) *DeleteResourceParser {
	return &DeleteResourceParser{resources: resources}
}

func (d *DeleteResourceParser) Parse(g *dag.AcyclicGraph) (s v1.Status) {
	util.CheckNotNil(g, "graph is nil")
	if len(g.Vertices()) == 0 {
		log.Infof("no vertices in dag when parsing deleted resources. dag:%s", json.Marshal2String(g))
		return v1.NewErrorStatusWithMsg(v1.InvalidArgument, "no vertices in dag when parsing deleted resources")
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
		rn, s := graph.NewResourceNode(key, resourceIndex[key], models.Delete)
		if v1.IsErr(s) {
			return s
		}
		rnID := rn.Hashcode().(string)

		if !g.HasVertex(rn) && manifestGraphMap[rnID] == nil {
			log.Infof("resource:%v not found in models. Mark as delete node", key)
			// we cannot delete this node if any node dependsOn this node
			for _, v := range priorDependsOn[rnID] {
				if manifestGraphMap[v] != nil {
					msg := fmt.Sprintf("%s dependson %s, cannot delete resource %s", v, rnID, rnID)
					return v1.NewErrorStatusWithMsg(v1.Internal, msg)
				}
			}
			g.Add(rn)
			g.Connect(dag.BasicEdge(root, rn))
		}

		// compute implicit and explicate dependencies
		refNodeKeys, s := updateDependencies(resource)
		if v1.IsErr(s) {
			return s
		}

		// always get the latest vertex in the g.
		rn = GetVertex(g, rn).(*graph.ResourceNode)
		s = LinkRefNodes(g, refNodeKeys, resourceIndex, rn, models.Delete, manifestGraphMap)
		if v1.IsErr(s) {
			return s
		}
	}
	if err = g.Validate(); err != nil {
		return v1.NewErrorStatusWithMsg(v1.IllegalManifest, "Found circle dependency in models."+err.Error())
	}
	g.TransitiveReduction()
	return s
}

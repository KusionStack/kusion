package parser

import (
	"fmt"

	"kusionstack.io/kusion/pkg/engine/operation/graph"
	"kusionstack.io/kusion/pkg/engine/terraform/dag"

	"kusionstack.io/kusion/pkg/engine/operation/types"

	"kusionstack.io/kusion/pkg/engine/models"

	"kusionstack.io/kusion/pkg/status"
)

type Parser interface {
	Parse(dag *dag.AcyclicGraph) status.Status
}

func LinkRefNodes(ag *dag.AcyclicGraph, refNodeKeys []string, resourceIndex map[string]*models.Resource,
	rn dag.Vertex, defaultAction types.ActionType, manifestGraphMap map[string]interface{},
) status.Status {
	if len(refNodeKeys) == 0 {
		return nil
	}
	for _, parentKey := range refNodeKeys {
		if resourceIndex[parentKey] == nil {
			return status.NewErrorStatusWithMsg(status.IllegalManifest,
				fmt.Sprintf("can't find resource by key:%s in models or state.", parentKey))
		}
		parentNode, s := graph.NewResourceNode(parentKey, resourceIndex[parentKey], defaultAction)
		if status.IsErr(s) {
			return s
		}
		baseNode, s := graph.NewBaseNode(parentKey)
		if status.IsErr(s) {
			return s
		}

		switch defaultAction {
		case types.Delete:
			// if the parent node is a deleteNode, we will add an edge from child node to parent node.
			// if parent node is not a deleteNode and manifestGraph contains parent node,
			// we will add an edge from parent node to child node
			if manifestGraphMap[parentKey] == nil {
				if ag.HasVertex(parentNode) {
					parentNode = GetVertex(ag, baseNode).(*graph.ResourceNode)
					ag.Connect(dag.BasicEdge(rn, parentNode))
				} else {
					ag.Add(parentNode)
					ag.Connect(dag.BasicEdge(rn, parentNode))
				}
			} else {
				parentNode = GetVertex(ag, baseNode).(*graph.ResourceNode)
				ag.Connect(dag.BasicEdge(parentNode, rn))
			}
		default:
			hasParent := ag.HasVertex(parentNode)
			if hasParent {
				parentNode = GetVertex(ag, baseNode).(*graph.ResourceNode)
				ag.Connect(dag.BasicEdge(parentNode, rn))
			} else {
				ag.Add(parentNode)
				ag.Connect(dag.BasicEdge(parentNode, rn))
			}
		}
	}

	return nil
}

func Deduplicate(refNodeKeys []string) []string {
	allKeys := make(map[string]bool)
	var res []string
	for _, item := range refNodeKeys {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			res = append(res, item)
		}
	}
	return res
}

func GetVertex(g *dag.AcyclicGraph, nv dag.NamedVertex) interface{} {
	vertices := g.Vertices()
	for i, vertex := range vertices {
		if vertex.(dag.NamedVertex).Name() == nv.Name() {
			return vertices[i]
		}
	}
	return nil
}

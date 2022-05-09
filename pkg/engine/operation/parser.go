package operation

import (
	"fmt"

	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/status"

	"github.com/hashicorp/terraform/dag"
)

type Parser interface {
	Parse(dag *dag.AcyclicGraph) status.Status
}

func LinkRefNodes(graph *dag.AcyclicGraph, refNodeKeys []string, resourceIndex map[string]*states.ResourceState,
	rn dag.Vertex, defaultAction ActionType, manifestGraphMap map[string]interface{}) status.Status {
	if len(refNodeKeys) == 0 {
		return nil
	}
	for _, parentKey := range refNodeKeys {
		if resourceIndex[parentKey] == nil {
			return status.NewErrorStatusWithMsg(status.IllegalManifest,
				fmt.Sprintf("can't find resource by key:%s in manifest or state.", parentKey))
		}
		parentNode := NewResourceNode(parentKey, resourceIndex[parentKey], defaultAction)

		switch defaultAction {
		case Delete:
			// if the parent node is a deleteNode, we will add an edge from child node to parent node.
			// if parent node is not a deleteNode and manifestGraph contains parent node,
			// we will add an edge from parent node to child node
			if manifestGraphMap[parentKey] == nil {
				if graph.HasVertex(parentNode) {
					parentNode = GetVertex(graph, &BaseNode{ID: parentKey}).(*ResourceNode)
					graph.Connect(dag.BasicEdge(rn, parentNode))
				} else {
					graph.Add(parentNode)
					graph.Connect(dag.BasicEdge(rn, parentNode))
				}
			} else {
				parentNode = GetVertex(graph, &BaseNode{ID: parentKey}).(*ResourceNode)
				graph.Connect(dag.BasicEdge(parentNode, rn))
			}
		default:
			hasParent := graph.HasVertex(parentNode)
			if hasParent {
				parentNode = GetVertex(graph, &BaseNode{ID: parentKey}).(*ResourceNode)
				graph.Connect(dag.BasicEdge(parentNode, rn))
			} else {
				graph.Add(parentNode)
				graph.Connect(dag.BasicEdge(parentNode, rn))
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

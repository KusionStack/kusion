package parser

import (
	"fmt"
	"reflect"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation/graph"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/third_party/terraform/dag"
)

type Parser interface {
	Parse(dag *dag.AcyclicGraph) v1.Status
}

func updateDependencies(resource *apiv1.Resource) ([]string, v1.Status) {
	// handle explicate dependency
	refNodeKeys := resource.DependsOn

	// handle implicit dependency
	v := reflect.ValueOf(resource.Attributes)
	implicitRefKeys, _, s := graph.ReplaceImplicitRef(v, nil, func(
		res map[string]*apiv1.Resource,
		ref string,
	) (reflect.Value, v1.Status) {
		// don't replace anything when parsing dependencies
		return reflect.ValueOf(ref), nil
	})
	if v1.IsErr(s) {
		return nil, s
	}
	refNodeKeys = append(refNodeKeys, implicitRefKeys...)

	// Deduplicate
	refNodeKeys = Deduplicate(refNodeKeys)
	resource.DependsOn = refNodeKeys
	return refNodeKeys, nil
}

func LinkRefNodes(
	ag *dag.AcyclicGraph,
	refNodeKeys []string,
	resourceIndex map[string]*apiv1.Resource,
	rn dag.Vertex,
	defaultAction models.ActionType,
	manifestGraphMap map[string]interface{},
) v1.Status {
	if len(refNodeKeys) == 0 {
		return nil
	}
	for _, parentKey := range refNodeKeys {
		if resourceIndex[parentKey] == nil {
			return v1.NewErrorStatusWithMsg(v1.IllegalManifest,
				fmt.Sprintf("can't find resource by key:%s in models or state.", parentKey))
		}
		parentNode, s := graph.NewResourceNode(parentKey, resourceIndex[parentKey], defaultAction)
		if v1.IsErr(s) {
			return s
		}
		baseNode, s := graph.NewBaseNode(parentKey)
		if v1.IsErr(s) {
			return s
		}

		switch defaultAction {
		case models.Delete:
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

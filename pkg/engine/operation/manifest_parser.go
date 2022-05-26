package operation

import (
	"fmt"
	"reflect"
	"strings"

	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/status"
	"kusionstack.io/kusion/pkg/util"
	"kusionstack.io/kusion/pkg/util/json"

	"github.com/hashicorp/terraform/dag"
)

type ManifestParser struct {
	manifest *models.Spec
}

func NewManifestParser(manifest *models.Spec) *ManifestParser {
	return &ManifestParser{manifest: manifest}
}

var _ Parser = (*ManifestParser)(nil)

const (
	ImplicitRefPrefix = "$kusion_path."
)

func (m *ManifestParser) Parse(graph *dag.AcyclicGraph) (s status.Status) {
	util.CheckNotNil(graph, "dag is nil")
	mf := m.manifest
	util.CheckNotNil(mf, "models is nil")
	if mf.Resources == nil {
		sprintf := fmt.Sprintf("no resources in models:%s", json.Marshal2String(mf))
		return status.NewBaseStatus(status.Warning, status.NotFound, sprintf)
	}

	root, err := graph.Root()
	util.CheckNotError(err, "get graph root node error")
	util.CheckNotNil(root, fmt.Sprintf("No root in this DAG:%s", json.Marshal2String(graph)))
	resourceIndex := mf.Resources.Index()
	for key, resourceState := range resourceIndex {
		rn := NewResourceNode(key, resourceIndex[key], Update)

		// add this resource to dag at first time
		if !graph.HasVertex(rn) {
			graph.Add(rn)
			graph.Connect(dag.BasicEdge(root, rn))
		} else {
			// always get the latest vertex in this graph otherwise you will get subtle mistake in walking this graph
			rn = GetVertex(graph, rn).(*ResourceNode)
			graph.Connect(dag.BasicEdge(root, rn))
		}
		// handle explicate dependency
		refNodeKeys := resourceState.DependsOn

		// handle implicit dependency
		v := reflect.ValueOf(resourceState.Attributes)
		implicitRefKeys, _, s := ParseImplicitRef(v, nil, func(map[string]*models.Resource, string) (reflect.Value, status.Status) {
			return v, nil
		})
		if status.IsErr(s) {
			return s
		}
		refNodeKeys = append(refNodeKeys, implicitRefKeys...)

		// Deduplicate
		refNodeKeys = Deduplicate(refNodeKeys)

		// linkRefNodes
		s = LinkRefNodes(graph, refNodeKeys, resourceIndex, rn, Update, nil)
		if status.IsErr(s) {
			return s
		}
	}

	if err = graph.Validate(); err != nil {
		return status.NewErrorStatusWithMsg(status.IllegalManifest, "Found circle dependency in models:"+err.Error())
	}
	graph.TransitiveReduction()
	return s
}

func ParseImplicitRef(v reflect.Value, resourceIndex map[string]*models.Resource,
	replaceFun func(resourceIndex map[string]*models.Resource, refPath string) (reflect.Value, status.Status),
) ([]string, reflect.Value, status.Status) {
	var result []string
	if !v.IsValid() {
		return nil, v, status.NewErrorStatusWithMsg(status.InvalidArgument, "invalid implicit reference")
	}

	switch v.Type().Kind() {
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return nil, v, nil
		}
		return ParseImplicitRef(v.Elem(), resourceIndex, replaceFun)
	case reflect.String:
		vStr := v.String()
		if strings.HasPrefix(vStr, ImplicitRefPrefix) {
			ref := strings.TrimPrefix(vStr, ImplicitRefPrefix)
			util.CheckArgument(len(ref) > 0,
				fmt.Sprintf("illegal implicit ref:%s. Implicit ref format: %sresourceKey.attribute", ref, ImplicitRefPrefix))
			split := strings.Split(ref, ".")
			result = append(result, split[0])
			log.Infof("add implicit ref:%s", split[0])
			// replace v with output
			tv, s := replaceFun(resourceIndex, ref)
			if status.IsErr(s) {
				return nil, v, s
			}
			v = tv
		}
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return nil, v, nil
		}

		vs := reflect.MakeSlice(v.Type(), 0, 0)

		for i := 0; i < v.Len(); i++ {
			ref, tv, s := ParseImplicitRef(v.Index(i), resourceIndex, replaceFun)
			if status.IsErr(s) {
				return nil, tv, s
			}
			vs = reflect.Append(vs, tv)
			if ref != nil {
				result = append(result, ref...)
			}
		}
		v = vs
	case reflect.Map:
		if v.Len() == 0 {
			return nil, v, nil
		}
		makeMap := reflect.MakeMap(v.Type())

		iter := v.MapRange()
		for iter.Next() {
			ref, tv, s := ParseImplicitRef(iter.Value(), resourceIndex, replaceFun)
			if status.IsErr(s) {
				return nil, tv, s
			}
			if ref != nil {
				result = append(result, ref...)
			}
			makeMap.SetMapIndex(iter.Key(), tv)
		}
		v = makeMap
	}
	return result, v, nil
}

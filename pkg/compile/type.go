package compile

import (
	"reflect"
	"strconv"
	"strings"

	kcl "kusionstack.io/kclvm-go"

	"kusionstack.io/kusion/pkg/util/yaml"
)

// The result of a KCL compilation
type CompileResult struct {
	Documents []kcl.KCLResult
}

// New a CompileResult by KCLResultList
func NewCompileResult(k *kcl.KCLResultList) *CompileResult {
	return &CompileResult{
		Documents: k.Slice(),
	}
}

// New a CompileResult by map array
func NewCompileResultByMapList(mapList []map[string]interface{}) *CompileResult {
	documents := []kcl.KCLResult{}
	for _, mapItem := range mapList {
		documents = append(documents, kcl.KCLResult(mapItem))
	}
	return &CompileResult{
		Documents: documents,
	}
}

func (c *CompileResult) YAMLString() string {
	documentList := []interface{}{}
	for _, document := range c.Documents {
		documentList = append(documentList, document)
	}
	return yaml.MergeToOneYAML(documentList...)
}

type Topology struct {
	Idc      string `json:"idc,omitempty"`
	Cluster  string `json:"cluster,omitempty"`
	Zone     string `json:"zone,omitempty"`
	Replicas int    `json:"replicas,omitempty"`
}

func NewTopology(idc, cluster, zone string, replicas int) *Topology {
	return &Topology{
		Idc:      idc,
		Cluster:  cluster,
		Zone:     zone,
		Replicas: replicas,
	}
}

// 'idc=eu95,cluster=sigma-eu95,zone=CZ00A' => *Topology{...}
func NewTopologyByString(topologyString string) *Topology {
	topologyString = strings.TrimSpace(topologyString)
	if topologyString == "" {
		return nil
	}
	kvs := strings.Split(topologyString, ",")
	m := map[string]string{}
	for _, kv := range kvs {
		mapping := strings.Split(kv, "=")
		if len(mapping) == 2 {
			m[mapping[0]] = mapping[1]
		} else {
			panic("invalid topology string " + topologyString)
		}
	}
	replicas, err := strconv.Atoi(m["replicas"])
	if err != nil {
		panic(err)
	}
	return NewTopology(m["idc"], m["cluster"], m["zone"], replicas)
}

// *Topology{...} => "idc=eu95,cluster=sigma-eu95,zone=RZ00A,replicas=1",
// t.String() is equivalent to t.BuildKey()
func (t *Topology) String() string {
	kvs := t.KeyValueStrings()
	return strings.Join(kvs, ",")
}

// *Topology{...} => "idc=eu95,cluster=sigma-eu95,zone=RZ00A,replicas=1",
// t.BuildKey() is equivalent to t.String()
func (t *Topology) BuildKey() string {
	return t.String()
}

// *Topology{...} => []string{"idc=eu95", "cluster=sigma-eu95", "zone=RZ00A", "replicas=1"}
func (t *Topology) KeyValueStrings() []string {
	kvs := []string{}
	rt := reflect.TypeOf(*t)
	v := reflect.ValueOf(*t)
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		value := v.Field(i)
		if !value.IsZero() {
			fieldNameLower := strings.ToLower(field.Name)
			switch value.Type().Kind() {
			case reflect.String:
				kvs = append(kvs, fieldNameLower+"="+value.String())
			case reflect.Int:
				kvs = append(kvs, fieldNameLower+"="+strconv.Itoa(int(value.Int())))
			}
		}
	}
	return kvs
}

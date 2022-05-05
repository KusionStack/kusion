package compile

import (
	"reflect"
	"strings"
	"testing"

	"kusionstack.io/KCLVM/kclvm-go/api/kcl"
)

func TestCompileResult_YAMLString(t *testing.T) {
	type fields struct {
		Documents []kcl.KCLResult
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "t1",
			fields: fields{
				Documents: []kcl.KCLResult{{"a": "b"}},
			},
			want: "a: b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CompileResult{
				Documents: tt.fields.Documents,
			}
			if got := strings.TrimSpace(c.YAMLString()); got != tt.want {
				t.Errorf("CompileResult.YAMLString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewTopology(t *testing.T) {
	type args struct {
		idc      string
		cluster  string
		zone     string
		replicas int
	}
	tests := []struct {
		name string
		args args
		want *Topology
	}{
		{
			name: "t1",
			args: args{
				idc:      "eu95",
				cluster:  "sigma-eu95",
				zone:     "RZ00A",
				replicas: 1,
			},
			want: &Topology{
				Idc:      "eu95",
				Cluster:  "sigma-eu95",
				Zone:     "RZ00A",
				Replicas: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTopology(tt.args.idc, tt.args.cluster, tt.args.zone, tt.args.replicas); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTopology() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewTopologyByString(t *testing.T) {
	type args struct {
		topologyString string
	}
	tests := []struct {
		name string
		args args
		want *Topology
	}{
		{
			name: "full topology",
			args: args{
				topologyString: "idc=eu95,cluster=sigma-eu95,zone=RZ00A,replicas=1",
			},
			want: &Topology{
				Idc:      "eu95",
				Cluster:  "sigma-eu95",
				Zone:     "RZ00A",
				Replicas: 1,
			},
		},
		{
			name: "empty topology",
			args: args{
				topologyString: "",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTopologyByString(tt.args.topologyString); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTopologyByString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTopology_String(t *testing.T) {
	type fields struct {
		Idc      string
		Cluster  string
		Zone     string
		Replicas int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "full string",
			fields: fields{
				Idc:      "eu95",
				Cluster:  "sigma-eu95",
				Zone:     "RZ00A",
				Replicas: 1,
			},
			want: "idc=eu95,cluster=sigma-eu95,zone=RZ00A,replicas=1",
		},
		{
			name:   "empty string",
			fields: fields{},
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Topology{
				Idc:      tt.fields.Idc,
				Cluster:  tt.fields.Cluster,
				Zone:     tt.fields.Zone,
				Replicas: tt.fields.Replicas,
			}
			if got := tr.String(); got != tt.want {
				t.Errorf("Topology.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTopology_BuildKey(t *testing.T) {
	type fields struct {
		Idc      string
		Cluster  string
		Zone     string
		Replicas int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "full string",
			fields: fields{
				Idc:      "eu95",
				Cluster:  "sigma-eu95",
				Zone:     "RZ00A",
				Replicas: 1,
			},
			want: "idc=eu95,cluster=sigma-eu95,zone=RZ00A,replicas=1",
		},
		{
			name:   "empty string",
			fields: fields{},
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Topology{
				Idc:      tt.fields.Idc,
				Cluster:  tt.fields.Cluster,
				Zone:     tt.fields.Zone,
				Replicas: tt.fields.Replicas,
			}
			if got := tr.BuildKey(); got != tt.want {
				t.Errorf("Topology.BuildKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTopology_KeyValueStrings(t *testing.T) {
	type fields struct {
		Idc      string
		Cluster  string
		Zone     string
		Replicas int
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "full",
			fields: fields{
				Idc:      "eu95",
				Cluster:  "sigma-eu95",
				Zone:     "RZ00A",
				Replicas: 1,
			},
			want: []string{"idc=eu95", "cluster=sigma-eu95", "zone=RZ00A", "replicas=1"},
		},
		{
			name:   "empty",
			fields: fields{},
			want:   []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Topology{
				Idc:      tt.fields.Idc,
				Cluster:  tt.fields.Cluster,
				Zone:     tt.fields.Zone,
				Replicas: tt.fields.Replicas,
			}
			if got := tr.KeyValueStrings(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Topology.KeyValueStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewCompileResultByMapList(t *testing.T) {
	type args struct {
		mapList []map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want *CompileResult
	}{
		{
			name: "t1",
			args: args{
				mapList: []map[string]interface{}{{"replicas": 1}},
			},
			want: &CompileResult{
				Documents: []kcl.KCLResult{map[string]interface{}{"replicas": 1}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCompileResultByMapList(tt.args.mapList); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCompileResultByMapList() = %v, want %v", got, tt.want)
			}
		})
	}
}

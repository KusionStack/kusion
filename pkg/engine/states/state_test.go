package states

import (
	"reflect"
	"testing"

	"kusionstack.io/kusion/pkg/version"
)

func TestNewState(t *testing.T) {
	tests := []struct {
		name string
		want *State
	}{
		{
			name: "t1",
			want: &State{
				KusionVersion: version.ReleaseVersion(),
				Version:       1,
				Resources:     []ResourceState{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewState(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceKey(t *testing.T) {
	tests := []struct {
		name          string
		want          string
		resourceState *ResourceState
	}{
		{
			name: "t1",
			want: "kusion_test",
			resourceState: &ResourceState{
				ID:         "kusion_test",
				Mode:       "managed",
				Attributes: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resourceState.ResourceKey(); got != tt.want {
				t.Errorf("ResourceKey() = %v, want = %v", got, tt.want)
			}
		})
	}
}

func TestResources_Index(t *testing.T) {
	tests := []struct {
		name string
		rs   Resources
		want map[string]*ResourceState
	}{
		{
			name: "t1",
			rs: []ResourceState{
				{
					Mode: "managed",
					ID:   "a",
				},
			},
			want: map[string]*ResourceState{
				"a": {
					Mode: "managed",
					ID:   "a",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rs.Index(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Index() = %v, want = %v", got, tt.want)
			}
		})
	}
}

func TestResources_Len(t *testing.T) {
	tests := []struct {
		name string
		rs   Resources
		want int
	}{
		{
			name: "t1",
			rs: []ResourceState{
				{
					Mode: "a",
					ID:   "c",
				},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rs.Len(); got != tt.want {
				t.Errorf("manifest.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResources_Swap(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		rs   Resources
		args args
	}{
		{
			name: "t1",
			rs: []ResourceState{
				{
					Mode: "test",
					ID:   "test",
				},
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.rs.Swap(tt.args.i, tt.args.j)
		})
	}
}

func TestResources_Less(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		rs   Resources
		args args
		want bool
	}{
		{
			name: "t1",
			rs: []ResourceState{
				{
					Mode: "managed",
					ID:   "a",
				},
				{
					Mode: "managed",
					ID:   "b",
				},
			},
			args: args{0, 1},
			want: true,
		},
		{
			name: "t2",
			rs: []ResourceState{
				{
					Mode: "a",
					ID:   "a",
				},
				{
					Mode: "b",
					ID:   "b",
				},
			},
			args: args{0, 1},
			want: true,
		},
		{
			name: "t3",
			rs: []ResourceState{
				{
					Mode: "a",
					ID:   "a",
				},
				{
					Mode: "a",
					ID:   "a",
				},
			},
			args: args{0, 1},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rs.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("manifest.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

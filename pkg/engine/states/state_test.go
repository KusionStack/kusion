package states

import (
	"reflect"
	"testing"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
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
				Resources:     []apiv1.Resource{},
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
		resourceState *apiv1.Resource
	}{
		{
			name: "t1",
			want: "kusion_test",
			resourceState: &apiv1.Resource{
				ID:         "kusion_test",
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
		rs   apiv1.Resources
		want map[string]*apiv1.Resource
	}{
		{
			name: "t1",
			rs: []apiv1.Resource{
				{
					ID: "a",
				},
			},
			want: map[string]*apiv1.Resource{
				"a": {
					ID: "a",
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
		rs   apiv1.Resources
		want int
	}{
		{
			name: "t1",
			rs: []apiv1.Resource{
				{
					ID: "c",
				},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rs.Len(); got != tt.want {
				t.Errorf("models.Len() = %v, want %v", got, tt.want)
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
		rs   apiv1.Resources
		args args
	}{
		{
			name: "t1",
			rs: []apiv1.Resource{
				{
					ID: "test",
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
		rs   apiv1.Resources
		args args
		want bool
	}{
		{
			name: "t1",
			rs: []apiv1.Resource{
				{
					ID: "a",
				},
				{
					ID: "b",
				},
			},
			args: args{0, 1},
			want: true,
		},
		{
			name: "t2",
			rs: []apiv1.Resource{
				{
					ID: "a",
				},
				{
					ID: "b",
				},
			},
			args: args{0, 1},
			want: true,
		},
		{
			name: "t3",
			rs: []apiv1.Resource{
				{
					ID: "a",
				},
				{
					ID: "a",
				},
			},
			args: args{0, 1},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rs.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("models.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

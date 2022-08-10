package backend

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty/cty"
	_ "kusionstack.io/kusion/pkg/engine/backend/init"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/engine/states/local"
)

func TestMergeConfig(t *testing.T) {
	type args struct {
		config   map[string]interface{}
		override map[string]interface{}
	}
	type want struct {
		content map[string]interface{}
	}

	tests := map[string]struct {
		args
		want
	}{
		"MergeConfig": {
			args: args{
				config: map[string]interface{}{
					"path": "kusion_state.json",
				},
				override: map[string]interface{}{
					"config": "kusion_config.json",
				},
			},
			want: want{
				content: map[string]interface{}{
					"path":   "kusion_state.json",
					"config": "kusion_config.json",
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mergeConfig := MergeConfig(tt.config, tt.override)
			if diff := cmp.Diff(tt.want.content, mergeConfig); diff != "" {
				t.Errorf("\nWrapMergeConfigFailed(...): -want message, +got message:\n%s", diff)
			}
		})
	}
}

func TestBackendFromConfig(t *testing.T) {
	type args struct {
		config   *Storage
		override BackendOps
	}
	type want struct {
		storage states.StateStorage
		err     error
	}
	tests := map[string]struct {
		args
		want
	}{
		"BackendFromConfig": {
			args: args{
				config: &Storage{
					Type: "local",
					Config: map[string]interface{}{
						"path": "kusion_state.json",
					},
				},
				override: BackendOps{
					Config: []string{
						"path=kusion_local.json",
					},
				},
			},
			want: want{
				storage: &local.FileSystemState{Path: "kusion_local.json"},
				err:     nil,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			storage, _ := BackendFromConfig(tt.config, tt.override, "./")
			if diff := cmp.Diff(tt.want.storage, storage); diff != "" {
				t.Errorf("\nWrapBackendFromConfigFailed(...): -want message, +got message:\n%s", diff)
			}
		})
	}
}

func TestValidBackendConfig(t *testing.T) {
	type args struct {
		config map[string]interface{}
		schema cty.Type
	}
	type want struct {
		errMsg string
	}
	tests := map[string]struct {
		args
		want
	}{
		"InValidBackendConfig": {
			args: args{
				config: map[string]interface{}{
					"kusionPath": "kusion_state.json",
				},
				schema: cty.Object(map[string]cty.Type{"path": cty.String}),
			},
			want: want{
				errMsg: "not support kusionPath in backend config",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := validBackendConfig(tt.config, tt.schema)
			if diff := cmp.Diff(tt.want.errMsg, err.Error()); diff != "" {
				t.Errorf("\nWrapvalidBackendConfigFailed(...): -want message, +got message:\n%s", diff)
			}
		})
	}
}

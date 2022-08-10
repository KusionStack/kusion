//go:build !arm64
// +build !arm64

package local

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"kusionstack.io/kusion/pkg/engine/states"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

var stateFile string

func TestMain(m *testing.M) {
	currentDir, _ := os.Getwd()
	stateFile = filepath.Join(currentDir, "testdata", "kusion_state.json")

	m.Run()
	os.Exit(0)
}

func TestNewFileSystemState(t *testing.T) {
	tests := []struct {
		name string
		want states.StateStorage
	}{
		{
			name: "t1",
			want: &FileSystemState{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFileSystemState(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFileSystemState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileSystemState_GetLatestState(t *testing.T) {
	type fields struct {
		Path string
	}
	type args struct {
		query *states.StateQuery
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *states.State
		wantErr bool
	}{
		{
			name: "t1",
			fields: fields{
				Path: stateFile,
			},
			args: args{
				query: &states.StateQuery{},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &FileSystemState{
				Path: tt.fields.Path,
			}
			got, err := s.GetLatestState(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileSystemState.GetLatestState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FileSystemState.GetLatestState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func FileSystemStateSetUp(t *testing.T) *FileSystemState {
	monkey.Patch(os.WriteFile, func(filename string, data []byte, perm fs.FileMode) error {
		return nil
	})
	monkey.Patch(os.Remove, func(name string) error {
		return nil
	})

	return &FileSystemState{Path: "kusion_state_filesystem.json"}
}

func TestFileSystemState(t *testing.T) {
	defer monkey.UnpatchAll()
	fileSystemState := FileSystemStateSetUp(t)

	state := &states.State{Tenant: "test_global_tenant", Project: "test_project", Stack: "test_env"}
	err := fileSystemState.Apply(state)
	assert.NoError(t, err)

	err = fileSystemState.Delete("kusion_state_filesystem.json")
	assert.NoError(t, err)
}

package local

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/engine/states"
)

var stateFile, stateFileForDelete, deprecatedStateFile, deprecatedStateFileForDelete string

func TestMain(m *testing.M) {
	currentDir, _ := os.Getwd()
	stateFile = filepath.Join(currentDir, "testdata/test_stack", KusionStateFileFile)
	stateFileForDelete = filepath.Join(currentDir, "testdata/test_stack_for_delete", KusionStateFileFile)
	deprecatedStateFile = filepath.Join(currentDir, "testdata/deprecated_test_stack", KusionStateFileFile)
	deprecatedStateFileForDelete = filepath.Join(currentDir, "testdata/deprecated_test_stack_for_delete", KusionStateFileFile)

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
		{
			name: "use deprecated kusion_state.json",
			fields: fields{
				Path: deprecatedStateFile,
			},
			args: args{
				query: &states.StateQuery{},
			},
			want:    &states.State{ID: 1},
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
	mockey.Mock(os.WriteFile).To(func(filename string, data []byte, perm fs.FileMode) error {
		return nil
	}).Build()
	mockey.Mock(os.Remove).To(func(name string) error {
		return nil
	}).Build()

	return &FileSystemState{Path: "kusion_state_filesystem.yaml"}
}

func TestFileSystemState(t *testing.T) {
	fileSystemState := FileSystemStateSetUp(t)

	state := &states.State{Tenant: "test_global_tenant", Project: "test_project", Stack: "test_env"}
	err := fileSystemState.Apply(state)
	assert.NoError(t, err)

	err = fileSystemState.Delete("kusion_state_filesystem.yaml")
	assert.NoError(t, err)
}

func TestFileSystem_Delete(t *testing.T) {
	testcases := []struct {
		name                   string
		success                bool
		stateFilePath          string
		useDeprecatedStateFile bool
	}{
		{
			name:                   "delete default state file",
			success:                true,
			stateFilePath:          stateFileForDelete,
			useDeprecatedStateFile: false,
		},
		{
			name:                   "delete both default and deprecated state file",
			success:                true,
			stateFilePath:          deprecatedStateFileForDelete,
			useDeprecatedStateFile: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			fileSystemState := &FileSystemState{Path: tc.stateFilePath}
			err := fileSystemState.Delete("")
			assert.NoError(t, err)
			assert.NoFileExists(t, tc.stateFilePath)
			file, _ := os.Create(tc.stateFilePath)
			_ = file.Close()
			if tc.useDeprecatedStateFile {
				dir := filepath.Dir(tc.stateFilePath)
				deprecatedStateFilePath := path.Join(dir, deprecatedKusionStateFile)
				assert.NoFileExists(t, deprecatedStateFilePath)
				deprecatedFile, _ := os.Create(deprecatedStateFilePath)
				_ = deprecatedFile.Close()
			}
		})
	}
}

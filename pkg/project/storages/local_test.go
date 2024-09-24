package storages

import (
	"os"
	"testing"
)

func TestNewLocalStorage(t *testing.T) {
	tests := []struct {
		name string
		path string
		want *LocalStorage
	}{
		{
			name: "NewLocalStorage with valid path",
			path: "/valid/path",
			want: &LocalStorage{path: "/valid/path"},
		},
		{
			name: "NewLocalStorage with empty path",
			path: "",
			want: &LocalStorage{path: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewLocalStorage(tt.path)
			if got.path != tt.want.path {
				t.Errorf("NewLocalStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGet(t *testing.T) {
	// Setup a LocalStorage with a temporary directory
	tempDir, err := os.MkdirTemp("", "test_local_storage")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // clean up

	localStorage := LocalStorage{path: tempDir}

	// Create dummy projects and workspaces
	dummyProjects := []string{"project1", "project2"}
	for _, projectName := range dummyProjects {
		projectPath := localStorage.path + "/" + projectName
		if err := os.Mkdir(projectPath, 0o755); err != nil {
			t.Fatalf("failed to create project dir: %v", err)
		}
		dummyWorkspaces := []string{"workspace1", "workspace2"}
		for _, workspaceName := range dummyWorkspaces {
			workspacePath := projectPath + "/" + workspaceName
			if err := os.Mkdir(workspacePath, 0o755); err != nil {
				t.Fatalf("failed to create workspace dir: %v", err)
			}
		}
	}

	// Call Get() method
	got, err := localStorage.Get()
	if err != nil {
		t.Errorf("LocalStorage.Get() error = %v, wantErr %v", err, false)
		return
	}

	// Check the result
	want := map[string][]string{
		"workspace1": {"project1", "project2"},
		"workspace2": {"project1", "project2"},
	}
	if len(got) != len(want) {
		t.Errorf("LocalStorage.Get() got = %v, want %v", got, want)
	}
	for k, v := range want {
		if !equalSlices(got[k], v) {
			t.Errorf("LocalStorage.Get() got = %v, want %v", got, want)
		}
	}
}

// equalSlices checks if two string slices are equal.
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

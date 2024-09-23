package storages

import (
	"fmt"
	"os"
)

// LocalStorage is an implementation of graph.Storage which uses local filesystem as storage.
type LocalStorage struct {
	// The directory path to store the project folders.
	path string
}

// NewLocalStorage creates a new LocalStorage instance.
func NewLocalStorage(path string) *LocalStorage {
	return &LocalStorage{path: path}
}

// Get returns a project map which key is workspace name and value is its belonged project list.
func (s *LocalStorage) Get() (map[string][]string, error) {
	projects := map[string][]string{}
	projectDir, err := os.ReadDir(s.path)
	if err != nil {
		return nil, fmt.Errorf("read releases folder failed: %w", err)
	}
	for _, project := range projectDir {
		if project.IsDir() {
			workspaces, err := os.ReadDir(fmt.Sprintf("%s/%s", s.path, project.Name()))
			if err != nil {
				return nil, fmt.Errorf("read workspace folder failed: %w", err)
			}
			for _, workspace := range workspaces {
				if workspace.IsDir() {
					workspaceName := workspace.Name()
					projects[workspaceName] = append(projects[workspaceName], project.Name())
				}
			}
		}
	}
	return projects, nil
}

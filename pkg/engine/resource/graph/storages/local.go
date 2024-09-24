package storages

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/engine/resource/graph"
)

// LocalStorage is an implementation of resource.Storage which uses local filesystem as storage.
type LocalStorage struct {
	// The directory path to store the resource files.
	path string
}

// NewLocalStorage news local resource storage, and derives graph.
// For instance, local path is ~/.kusion/resources/project/workspace
func NewLocalStorage(path string) (*LocalStorage, error) {
	s := &LocalStorage{path: path}

	// create the resources directory
	if err := os.MkdirAll(s.path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("create resources directory failed, %w", err)
	}
	return s, nil
}

// Get gets the graph from local.
func (s *LocalStorage) Get() (*v1.Graph, error) {
	content, err := os.ReadFile(filepath.Join(s.path, graphFileName))
	if err != nil {
		return nil, fmt.Errorf("read resource graph file failed: %w", err)
	}

	r := &v1.Graph{}
	if err = json.Unmarshal(content, r); err != nil {
		return nil, fmt.Errorf("json unmarshal graph failed: %w", err)
	}

	// Index is not stored in s3, so we need to rebuild it.
	// Update resource index to use index in the memory.
	graph.UpdateResourceIndex(r.Resources)

	return r, nil
}

// Create creates the graph in s3.
func (s *LocalStorage) Create(r *v1.Graph) error {
	content, _ := os.ReadFile(filepath.Join(s.path, graphFileName))
	if content != nil {
		return ErrGraphAlreadyExist
	}

	return s.writeGraph(r)
}

// Update updates the graph in s3.
func (s *LocalStorage) Update(r *v1.Graph) error {
	_, err := os.ReadFile(filepath.Join(s.path, graphFileName))
	if err != nil {
		return ErrGraphNotExist
	}

	return s.writeGraph(r)
}

// Delete deletes the graph in s3
func (s *LocalStorage) Delete() error {
	_, err := os.ReadFile(filepath.Join(s.path, graphFileName))
	if !os.IsNotExist(err) {
		if err := os.Remove(filepath.Join(s.path, graphFileName)); err != nil {
			return fmt.Errorf("remove graph file failed: %w", err)
		}
	}

	return nil
}

// writeGraph writes the graph to s3.
func (s *LocalStorage) writeGraph(r *v1.Graph) error {
	content, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("json marshal graph failed: %w", err)
	}

	if err = os.WriteFile(filepath.Join(s.path, graphFileName), content, os.ModePerm); err != nil {
		return fmt.Errorf("write graph file failed: %w", err)
	}

	return nil
}

// CheckGraphStorageExistence checks whether the graph storage exists.
func (s *LocalStorage) CheckGraphStorageExistence() bool {
	_, err := os.ReadFile(filepath.Join(s.path, graphFileName))
	return !os.IsNotExist(err)
}

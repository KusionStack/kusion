package storages

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// LocalStorage is an implementation of workspace.Storage which uses local filesystem as storage.
type LocalStorage struct {
	// The directory path to store the workspace files.
	path string

	meta *workspacesMetaData
}

// NewLocalStorage news local workspace storage and init default workspace.
func NewLocalStorage(path string) (*LocalStorage, error) {
	s := &LocalStorage{path: path}

	// create the workspace directory
	if err := os.MkdirAll(s.path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("create workspace directory failed, %w", err)
	}
	// read workspaces metadata
	if err := s.readMeta(); err != nil {
		return nil, err
	}

	return s, s.initDefaultWorkspaceIf()
}

func (s *LocalStorage) Get(name string) (*v1.Workspace, error) {
	if name == "" {
		name = s.meta.Current
	}
	if !checkWorkspaceExistence(s.meta, name) {
		return nil, ErrWorkspaceNotExist
	}

	content, err := os.ReadFile(filepath.Join(s.path, name+yamlSuffix))
	if err != nil {
		return nil, fmt.Errorf("read workspace file failed: %w", err)
	}

	ws := &v1.Workspace{}
	if err = yaml.Unmarshal(content, ws); err != nil {
		return nil, fmt.Errorf("yaml unmarshal workspace failed: %w", err)
	}
	ws.Name = name
	return ws, nil
}

func (s *LocalStorage) Create(ws *v1.Workspace) error {
	if checkWorkspaceExistence(s.meta, ws.Name) {
		return ErrWorkspaceAlreadyExist
	}

	if err := s.writeWorkspace(ws); err != nil {
		return err
	}

	addAvailableWorkspaces(s.meta, ws.Name)
	return s.writeMeta()
}

func (s *LocalStorage) Update(ws *v1.Workspace) error {
	if ws.Name == "" {
		ws.Name = s.meta.Current
	}
	if !checkWorkspaceExistence(s.meta, ws.Name) {
		return ErrWorkspaceNotExist
	}

	return s.writeWorkspace(ws)
}

func (s *LocalStorage) Delete(name string) error {
	if name == "" {
		name = s.meta.Current
	}
	if !checkWorkspaceExistence(s.meta, name) {
		return nil
	}

	if err := os.Remove(filepath.Join(s.path, name+yamlSuffix)); err != nil {
		return fmt.Errorf("remove workspace file failed: %w", err)
	}

	removeAvailableWorkspaces(s.meta, name)
	return s.writeMeta()
}

func (s *LocalStorage) GetNames() ([]string, error) {
	return s.meta.AvailableWorkspaces, nil
}

func (s *LocalStorage) GetCurrent() (string, error) {
	return s.meta.Current, nil
}

func (s *LocalStorage) SetCurrent(name string) error {
	if !checkWorkspaceExistence(s.meta, name) {
		return ErrWorkspaceNotExist
	}
	s.meta.Current = name
	return s.writeMeta()
}

func (s *LocalStorage) initDefaultWorkspaceIf() error {
	if !checkWorkspaceExistence(s.meta, DefaultWorkspace) {
		// if there is no default workspace, create one with empty workspace.
		if err := s.writeWorkspace(&v1.Workspace{Name: DefaultWorkspace}); err != nil {
			return err
		}
		addAvailableWorkspaces(s.meta, DefaultWorkspace)
	}

	if s.meta.Current == "" {
		s.meta.Current = DefaultWorkspace
	}
	return s.writeMeta()
}

func (s *LocalStorage) readMeta() error {
	content, err := os.ReadFile(filepath.Join(s.path, metadataFile))
	if os.IsNotExist(err) {
		s.meta = &workspacesMetaData{}
		return nil
	} else if err != nil {
		return fmt.Errorf("read workspace metadata file failed: %w", err)
	}

	meta := &workspacesMetaData{}
	if err = yaml.Unmarshal(content, meta); err != nil {
		return fmt.Errorf("yaml unmarshal workspaces metadata failed: %w", err)
	}
	s.meta = meta
	return nil
}

func (s *LocalStorage) writeMeta() error {
	content, err := yaml.Marshal(s.meta)
	if err != nil {
		return fmt.Errorf("yaml marshal workspaces metadata failed: %w", err)
	}

	if err = os.WriteFile(filepath.Join(s.path, metadataFile), content, os.ModePerm); err != nil {
		return fmt.Errorf("write workspaces metadata file failed: %w", err)
	}
	return nil
}

func (s *LocalStorage) writeWorkspace(ws *v1.Workspace) error {
	content, err := yaml.Marshal(ws)
	if err != nil {
		return fmt.Errorf("yaml marshal workspace failed: %w", err)
	}

	if err = os.WriteFile(filepath.Join(s.path, ws.Name+yamlSuffix), content, os.ModePerm); err != nil {
		return fmt.Errorf("write workspace file failed: %w", err)
	}
	return nil
}

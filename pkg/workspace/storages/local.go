package storages

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

// LocalStorage is an implementation of workspace.Storage which uses local filesystem as storage.
type LocalStorage struct {
	// The directory path to store the workspace files.
	path string
}

// NewLocalStorage news local workspace storage and init default workspace.
func NewLocalStorage(path string) (*LocalStorage, error) {
	s := &LocalStorage{path: path}

	// create the workspace directory
	if err := os.MkdirAll(s.path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("create workspace directory failed, %w", err)
	}

	return s, s.initDefaultWorkspaceIf()
}

func (s *LocalStorage) Get(name string) (*v1.Workspace, error) {
	exist, err := s.Exist(name)
	if err != nil {
		return nil, err
	}
	if !exist {
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
	meta, err := s.readMeta()
	if err != nil {
		return err
	}
	if checkWorkspaceExistence(meta, ws.Name) {
		return ErrWorkspaceAlreadyExist
	}

	if err = s.writeWorkspace(ws); err != nil {
		return err
	}

	addAvailableWorkspaces(meta, ws.Name)
	return s.writeMeta(meta)
}

func (s *LocalStorage) Update(ws *v1.Workspace) error {
	exist, err := s.Exist(ws.Name)
	if err != nil {
		return err
	}
	if !exist {
		return ErrWorkspaceNotExist
	}

	return s.writeWorkspace(ws)
}

func (s *LocalStorage) Delete(name string) error {
	meta, err := s.readMeta()
	if err != nil {
		return err
	}
	if !checkWorkspaceExistence(meta, name) {
		return nil
	}

	if err = os.Remove(filepath.Join(s.path, name+yamlSuffix)); err != nil {
		return fmt.Errorf("remove workspace file failed: %w", err)
	}

	removeAvailableWorkspaces(meta, name)
	return s.writeMeta(meta)
}

func (s *LocalStorage) Exist(name string) (bool, error) {
	meta, err := s.readMeta()
	if err != nil {
		return false, err
	}
	return checkWorkspaceExistence(meta, name), nil
}

func (s *LocalStorage) GetNames() ([]string, error) {
	meta, err := s.readMeta()
	if err != nil {
		return nil, err
	}
	return meta.AvailableWorkspaces, nil
}

func (s *LocalStorage) GetCurrent() (string, error) {
	meta, err := s.readMeta()
	if err != nil {
		return "", err
	}
	return meta.Current, nil
}

func (s *LocalStorage) SetCurrent(name string) error {
	meta, err := s.readMeta()
	if err != nil {
		return err
	}
	if !checkWorkspaceExistence(meta, name) {
		return ErrWorkspaceNotExist
	}
	meta.Current = name
	return s.writeMeta(meta)
}

func (s *LocalStorage) initDefaultWorkspaceIf() error {
	meta, err := s.readMeta()
	if err != nil {
		return err
	}
	if !checkWorkspaceExistence(meta, defaultWorkspace) {
		// if there is no default workspace, create one with empty workspace.
		if err = s.writeWorkspace(&v1.Workspace{Name: defaultWorkspace}); err != nil {
			return err
		}
		addAvailableWorkspaces(meta, defaultWorkspace)
	}

	if meta.Current == "" {
		meta.Current = defaultWorkspace
	}
	return s.writeMeta(meta)
}

func (s *LocalStorage) readMeta() (*workspacesMetaData, error) {
	content, err := os.ReadFile(filepath.Join(s.path, metadataFile))
	if os.IsNotExist(err) {
		return &workspacesMetaData{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("read workspace meta data file failed: %w", err)
	}

	meta := &workspacesMetaData{}
	if err = yaml.Unmarshal(content, meta); err != nil {
		return nil, fmt.Errorf("yaml unmarshal workspaces meta data failed: %w", err)
	}
	return meta, nil
}

func (s *LocalStorage) writeMeta(meta *workspacesMetaData) error {
	content, err := yaml.Marshal(meta)
	if err != nil {
		return fmt.Errorf("yaml marshal workspaces meta data failed: %w", err)
	}

	if err = os.WriteFile(filepath.Join(s.path, metadataFile), content, os.ModePerm); err != nil {
		return fmt.Errorf("write workspaces meta data file failed: %w", err)
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

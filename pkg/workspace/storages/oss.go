package storages

import (
	"bytes"
	"fmt"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// OssStorage is an implementation of workspace.Storage which uses oss as storage.
type OssStorage struct {
	bucket *oss.Bucket

	// The prefix to store the workspaces files.
	prefix string

	meta *workspacesMetaData
}

// NewOssStorage news oss workspace storage and init default workspace.
func NewOssStorage(bucket *oss.Bucket, prefix string) (*OssStorage, error) {
	s := &OssStorage{
		bucket: bucket,
		prefix: prefix,
	}
	if err := s.readMeta(); err != nil {
		return nil, err
	}
	return s, s.initDefaultWorkspaceIf()
}

func (s *OssStorage) Get(name string) (*v1.Workspace, error) {
	if name == "" {
		name = s.meta.Current
	}
	if !checkWorkspaceExistence(s.meta, name) {
		return nil, ErrWorkspaceNotExist
	}

	body, err := s.bucket.GetObject(s.prefix + "/" + name + yamlSuffix)
	if err != nil {
		return nil, fmt.Errorf("get workspace from oss failed: %w", err)
	}
	defer func() {
		_ = body.Close()
	}()
	content, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read workspace failed: %w", err)
	}

	ws := &v1.Workspace{}
	if err = yaml.Unmarshal(content, ws); err != nil {
		return nil, fmt.Errorf("yaml unmarshal workspace failed: %w", err)
	}
	ws.Name = name
	return ws, nil
}

func (s *OssStorage) Create(ws *v1.Workspace) error {
	if checkWorkspaceExistence(s.meta, ws.Name) {
		return ErrWorkspaceAlreadyExist
	}

	if err := s.writeWorkspace(ws); err != nil {
		return err
	}

	addAvailableWorkspaces(s.meta, ws.Name)
	return s.writeMeta()
}

func (s *OssStorage) Update(ws *v1.Workspace) error {
	if ws.Name == "" {
		ws.Name = s.meta.Current
	}
	if !checkWorkspaceExistence(s.meta, ws.Name) {
		return ErrWorkspaceNotExist
	}

	return s.writeWorkspace(ws)
}

func (s *OssStorage) Delete(name string) error {
	if name == "" {
		name = s.meta.Current
	}
	if !checkWorkspaceExistence(s.meta, name) {
		return nil
	}

	if err := s.bucket.DeleteObject(s.prefix + "/" + name + yamlSuffix); err != nil {
		return fmt.Errorf("remove workspace in oss failed: %w", err)
	}

	removeAvailableWorkspaces(s.meta, name)
	return s.writeMeta()
}

func (s *OssStorage) GetNames() ([]string, error) {
	return s.meta.AvailableWorkspaces, nil
}

func (s *OssStorage) GetCurrent() (string, error) {
	return s.meta.Current, nil
}

func (s *OssStorage) SetCurrent(name string) error {
	if !checkWorkspaceExistence(s.meta, name) {
		return ErrWorkspaceNotExist
	}
	s.meta.Current = name
	return s.writeMeta()
}

func (s *OssStorage) RenameWorkspace(oldName, newName string) (err error) {
	if oldName == "" || newName == "" {
		return fmt.Errorf("given name is empty")
	}

	// restore the old workspace name if the rename failed
	defer func() {
		if err != nil {
			removeAvailableWorkspaces(s.meta, newName)
			addAvailableWorkspaces(s.meta, oldName)
			s.writeMeta()
		}
	}()

	// update the meta file
	removeAvailableWorkspaces(s.meta, oldName)
	addAvailableWorkspaces(s.meta, newName)
	if err = s.writeMeta(); err != nil {
		return err
	}

	// rename the workspace file by copying and deleting
	oldPath := s.prefix + "/" + oldName + yamlSuffix
	newPath := s.prefix + "/" + newName + yamlSuffix

	// get old workspace content
	body, err := s.bucket.GetObject(oldPath)
	if err != nil {
		return fmt.Errorf("get old workspace failed: %w", err)
	}
	defer body.Close()

	// copy to new path
	if err = s.bucket.PutObject(newPath, body); err != nil {
		return fmt.Errorf("copy workspace failed: %w", err)
	}

	// delete old file
	if err = s.bucket.DeleteObject(oldPath); err != nil {
		return fmt.Errorf("delete old workspace failed: %w", err)
	}

	return nil
}

func (s *OssStorage) initDefaultWorkspaceIf() error {
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

func (s *OssStorage) readMeta() error {
	body, err := s.bucket.GetObject(s.prefix + "/" + metadataFile)
	if err != nil {
		ossErr, ok := err.(oss.ServiceError)
		// error code ref: github.com/aliyun/aliyun-oss-go-sdk@v2.1.8+incompatible/oss/bucket.go:553
		if ok && ossErr.StatusCode == 404 {
			s.meta = &workspacesMetaData{}
			return nil
		}
		return fmt.Errorf("get workspaces metadata from oss failed: %w", err)
	}
	defer func() {
		_ = body.Close()
	}()

	content, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("read workspaces metadata failed: %w", err)
	}
	if len(content) == 0 {
		s.meta = &workspacesMetaData{}
		return nil
	}

	meta := &workspacesMetaData{}
	if err = yaml.Unmarshal(content, meta); err != nil {
		return fmt.Errorf("yaml unmarshal workspaces metadata failed: %w", err)
	}
	s.meta = meta
	return nil
}

func (s *OssStorage) writeMeta() error {
	content, err := yaml.Marshal(s.meta)
	if err != nil {
		return fmt.Errorf("yaml marshal workspaces metadata failed: %w", err)
	}

	if err = s.bucket.PutObject(s.prefix+"/"+metadataFile, bytes.NewReader(content)); err != nil {
		return fmt.Errorf("put workspaces metadata to oss failed: %w", err)
	}
	return nil
}

func (s *OssStorage) writeWorkspace(ws *v1.Workspace) error {
	content, err := yaml.Marshal(ws)
	if err != nil {
		return fmt.Errorf("yaml marshal workspace failed: %w", err)
	}

	if err = s.bucket.PutObject(s.prefix+"/"+ws.Name+yamlSuffix, bytes.NewReader(content)); err != nil {
		return fmt.Errorf("put workspace to oss failed: %w", err)
	}
	return nil
}

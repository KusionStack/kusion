package storages

import (
	"bytes"
	"fmt"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

// OssStorage is an implementation of workspace.Storage which uses oss as storage.
type OssStorage struct {
	bucket *oss.Bucket

	// The prefix to store the workspaces files.
	prefix string
}

// NewOssStorage news oss workspace storage and init default workspace.
func NewOssStorage(bucket *oss.Bucket, prefix string) (*OssStorage, error) {
	s := &OssStorage{
		bucket: bucket,
		prefix: prefix,
	}
	return s, s.initDefaultWorkspaceIf()
}

func (s *OssStorage) Get(name string) (*v1.Workspace, error) {
	exist, err := s.Exist(name)
	if err != nil {
		return nil, err
	}
	if !exist {
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

func (s *OssStorage) Update(ws *v1.Workspace) error {
	exist, err := s.Exist(ws.Name)
	if err != nil {
		return err
	}
	if !exist {
		return ErrWorkspaceNotExist
	}

	return s.writeWorkspace(ws)
}

func (s *OssStorage) Delete(name string) error {
	meta, err := s.readMeta()
	if err != nil {
		return err
	}
	if !checkWorkspaceExistence(meta, name) {
		return nil
	}

	if err = s.bucket.DeleteObject(s.prefix + "/" + name + yamlSuffix); err != nil {
		return fmt.Errorf("remove workspace in oss failed: %w", err)
	}

	removeAvailableWorkspaces(meta, name)
	return s.writeMeta(meta)
}

func (s *OssStorage) Exist(name string) (bool, error) {
	meta, err := s.readMeta()
	if err != nil {
		return false, err
	}
	return checkWorkspaceExistence(meta, name), nil
}

func (s *OssStorage) GetNames() ([]string, error) {
	meta, err := s.readMeta()
	if err != nil {
		return nil, err
	}
	return meta.AvailableWorkspaces, nil
}

func (s *OssStorage) GetCurrent() (string, error) {
	meta, err := s.readMeta()
	if err != nil {
		return "", err
	}
	return meta.Current, nil
}

func (s *OssStorage) SetCurrent(name string) error {
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

func (s *OssStorage) initDefaultWorkspaceIf() error {
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

func (s *OssStorage) readMeta() (*workspacesMetaData, error) {
	body, err := s.bucket.GetObject(s.prefix + "/" + metadataFile)
	if err != nil {
		ossErr, ok := err.(oss.ServiceError)
		// error code ref: github.com/aliyun/aliyun-oss-go-sdk@v2.1.8+incompatible/oss/bucket.go:553
		if ok && ossErr.StatusCode == 404 {
			return &workspacesMetaData{}, nil
		}
		return nil, fmt.Errorf("get workspaces meta data from oss failed: %w", err)
	}
	defer func() {
		_ = body.Close()
	}()

	content, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read workspaces meta data failed: %w", err)
	}
	if len(content) == 0 {
		return &workspacesMetaData{}, nil
	}

	meta := &workspacesMetaData{}
	if err = yaml.Unmarshal(content, meta); err != nil {
		return nil, fmt.Errorf("yaml unmarshal workspaces meta data failed: %w", err)
	}
	return meta, nil
}

func (s *OssStorage) writeMeta(meta *workspacesMetaData) error {
	content, err := yaml.Marshal(meta)
	if err != nil {
		return fmt.Errorf("yaml marshal workspaces meta data failed: %w", err)
	}

	if err = s.bucket.PutObject(s.prefix+"/"+metadataFile, bytes.NewReader(content)); err != nil {
		return fmt.Errorf("put workspaces meta data to oss failed: %w", err)
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

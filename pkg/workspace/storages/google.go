package storages

import (
	"context"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"

	googlestorage "cloud.google.com/go/storage"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// GoogleStorage is an implementation of workspace.Storage which uses google cloud as storage.
type GoogleStorage struct {
	bucket googlestorage.BucketHandle

	// The prefix to store the workspaces files.
	prefix string

	meta *workspacesMetaData
}

// NewGoogleStorage news google cloud workspace storage and init default workspace.
func NewGoogleStorage(bucket *googlestorage.BucketHandle, prefix string) (*GoogleStorage, error) {
	s := &GoogleStorage{
		bucket: *bucket,
		prefix: prefix,
	}
	if err := s.readMeta(); err != nil {
		return nil, err
	}
	return s, s.initDefaultWorkspaceIf()
}

func (s *GoogleStorage) Get(name string) (*v1.Workspace, error) {
	if name == "" {
		name = s.meta.Current
	}
	if !checkWorkspaceExistence(s.meta, name) {
		return nil, ErrWorkspaceNotExist
	}

	obj := s.bucket.Object(s.prefix + "/" + name + yamlSuffix)
	reader, err := obj.NewReader(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get workspace from google storage failed: %w", err)
	}
	defer reader.Close()
	content, err := io.ReadAll(reader)
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

func (s *GoogleStorage) Create(ws *v1.Workspace) error {
	if checkWorkspaceExistence(s.meta, ws.Name) {
		return ErrWorkspaceAlreadyExist
	}

	if err := s.writeWorkspace(ws); err != nil {
		return err
	}

	addAvailableWorkspaces(s.meta, ws.Name)
	return s.writeMeta()
}

func (s *GoogleStorage) Update(ws *v1.Workspace) error {
	if ws.Name == "" {
		ws.Name = s.meta.Current
	}
	if !checkWorkspaceExistence(s.meta, ws.Name) {
		return ErrWorkspaceNotExist
	}

	return s.writeWorkspace(ws)
}

func (s *GoogleStorage) Delete(name string) error {
	if !checkWorkspaceExistence(s.meta, name) {
		return nil
	}

	obj := s.bucket.Object(s.prefix + "/" + name + yamlSuffix)
	if err := obj.Delete(context.Background()); err != nil {
		return fmt.Errorf("remove workspace in google storage failed: %w", err)
	}

	removeAvailableWorkspaces(s.meta, name)
	return s.writeMeta()
}

func (s *GoogleStorage) GetNames() ([]string, error) {
	return s.meta.AvailableWorkspaces, nil
}

func (s *GoogleStorage) GetCurrent() (string, error) {
	return s.meta.Current, nil
}

func (s *GoogleStorage) SetCurrent(name string) error {
	if !checkWorkspaceExistence(s.meta, name) {
		return ErrWorkspaceNotExist
	}
	s.meta.Current = name
	return s.writeMeta()
}

func (s *GoogleStorage) RenameWorkspace(oldName, newName string) (err error) {
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

	// rename the workspace file
	oldObj := s.bucket.Object(s.prefix + "/" + oldName + yamlSuffix)
	dstObj := s.bucket.Object(s.prefix + "/" + newName + yamlSuffix)
	if _, err = dstObj.CopierFrom(oldObj).Run(context.Background()); err != nil {
		return fmt.Errorf("rename workspace file failed: %w", err)
	}
	if err = oldObj.Delete(context.Background()); err != nil {
		return fmt.Errorf("delete old workspace file failed: %w", err)
	}

	return nil
}

func (s *GoogleStorage) initDefaultWorkspaceIf() error {
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

func (s *GoogleStorage) readMeta() error {
	ctx := context.Background()
	obj := s.bucket.Object(s.prefix + "/" + metadataFile)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		if err == googlestorage.ErrObjectNotExist {
			s.meta = &workspacesMetaData{}
			return nil
		}
		return fmt.Errorf("get workspaces metadata from google failed: %w", err)
	}
	defer reader.Close()
	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read workspaces meta data failed: %w", err)
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

func (s *GoogleStorage) writeMeta() error {
	ctx := context.Background()
	obj := s.bucket.Object(s.prefix + "/" + metadataFile)
	content, err := yaml.Marshal(s.meta)
	if err != nil {
		return fmt.Errorf("yaml marshal workspaces metadata failed: %w", err)
	}

	writer := obj.NewWriter(ctx)
	if _, err = writer.Write(content); err != nil {
		return fmt.Errorf("write workspaces metadata failed: %w", err)
	}

	if err = writer.Close(); err != nil {
		return fmt.Errorf("close writer failed: %w", err)
	}
	return nil
}

func (s *GoogleStorage) writeWorkspace(ws *v1.Workspace) error {
	content, err := yaml.Marshal(ws)
	if err != nil {
		return fmt.Errorf("yaml marshal workspace failed: %w", err)
	}

	obj := s.bucket.Object(s.prefix + "/" + ws.Name + yamlSuffix)
	writer := obj.NewWriter(context.Background())
	if _, err = writer.Write(content); err != nil {
		return fmt.Errorf("write workspace failed: %w", err)
	}

	if err = writer.Close(); err != nil {
		return fmt.Errorf("close writer failed: %w", err)
	}
	return nil
}

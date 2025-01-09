package storages

import (
	"bytes"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// S3Storage is an implementation of workspace.Storage which uses s3 as storage.
type S3Storage struct {
	s3     *s3.S3
	bucket string

	// The prefix to store the workspaces files.
	prefix string

	meta *workspacesMetaData
}

// NewS3Storage news s3 workspace storage and init default workspace.
func NewS3Storage(s3 *s3.S3, bucket, prefix string) (*S3Storage, error) {
	s := &S3Storage{
		s3:     s3,
		bucket: bucket,
		prefix: prefix,
	}
	if err := s.readMeta(); err != nil {
		return nil, err
	}
	return s, s.initDefaultWorkspaceIf()
}

func (s *S3Storage) Get(name string) (*v1.Workspace, error) {
	if name == "" {
		name = s.meta.Current
	}
	if !checkWorkspaceExistence(s.meta, name) {
		return nil, ErrWorkspaceNotExist
	}

	key := s.prefix + "/" + name + yamlSuffix
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    &key,
	}
	output, err := s.s3.GetObject(input)
	if err != nil {
		return nil, fmt.Errorf("get workspace from s3 failed: %w", err)
	}
	defer func() {
		_ = output.Body.Close()
	}()
	content, err := io.ReadAll(output.Body)
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

func (s *S3Storage) Create(ws *v1.Workspace) error {
	if checkWorkspaceExistence(s.meta, ws.Name) {
		return ErrWorkspaceAlreadyExist
	}

	if err := s.writeWorkspace(ws); err != nil {
		return err
	}

	addAvailableWorkspaces(s.meta, ws.Name)
	return s.writeMeta()
}

func (s *S3Storage) Update(ws *v1.Workspace) error {
	if ws.Name == "" {
		ws.Name = s.meta.Current
	}
	if !checkWorkspaceExistence(s.meta, ws.Name) {
		return ErrWorkspaceNotExist
	}

	return s.writeWorkspace(ws)
}

func (s *S3Storage) Delete(name string) error {
	if !checkWorkspaceExistence(s.meta, name) {
		return nil
	}

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.prefix + "/" + name + yamlSuffix),
	}
	if _, err := s.s3.DeleteObject(input); err != nil {
		return fmt.Errorf("remove workspace in s3 failed: %w", err)
	}

	removeAvailableWorkspaces(s.meta, name)
	return s.writeMeta()
}

func (s *S3Storage) GetNames() ([]string, error) {
	return s.meta.AvailableWorkspaces, nil
}

func (s *S3Storage) GetCurrent() (string, error) {
	return s.meta.Current, nil
}

func (s *S3Storage) SetCurrent(name string) error {
	if !checkWorkspaceExistence(s.meta, name) {
		return ErrWorkspaceNotExist
	}
	s.meta.Current = name
	return s.writeMeta()
}

func (s *S3Storage) RenameWorkspace(oldName, newName string) (err error) {
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

	// copy to new path
	copyInput := &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		CopySource: aws.String(s.bucket + "/" + oldPath),
		Key:        aws.String(newPath),
	}
	if _, err = s.s3.CopyObject(copyInput); err != nil {
		return fmt.Errorf("copy workspace failed: %w", err)
	}

	// delete old file
	deleteInput := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(oldPath),
	}
	if _, err = s.s3.DeleteObject(deleteInput); err != nil {
		return fmt.Errorf("delete old workspace failed: %w", err)
	}

	return nil
}

func (s *S3Storage) initDefaultWorkspaceIf() error {
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

func (s *S3Storage) readMeta() error {
	key := s.prefix + "/" + metadataFile
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    &key,
	}
	output, err := s.s3.GetObject(input)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == s3.ErrCodeNoSuchKey {
			s.meta = &workspacesMetaData{}
			return nil
		}
		return fmt.Errorf("get workspaces meta data from s3 failed: %w", err)
	}
	defer func() {
		_ = output.Body.Close()
	}()

	content, err := io.ReadAll(output.Body)
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

func (s *S3Storage) writeMeta() error {
	content, err := yaml.Marshal(s.meta)
	if err != nil {
		return fmt.Errorf("yaml marshal workspaces metadata failed: %w", err)
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.prefix + "/" + metadataFile),
		Body:   bytes.NewReader(content),
	}
	if _, err = s.s3.PutObject(input); err != nil {
		return fmt.Errorf("put workspaces metadata to s3 failed: %w", err)
	}
	return nil
}

func (s *S3Storage) writeWorkspace(ws *v1.Workspace) error {
	content, err := yaml.Marshal(ws)
	if err != nil {
		return fmt.Errorf("yaml marshal workspace failed: %w", err)
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.prefix + "/" + ws.Name + yamlSuffix),
		Body:   bytes.NewReader(content),
	}
	if _, err = s.s3.PutObject(input); err != nil {
		return fmt.Errorf("put workspace to s3 failed: %w", err)
	}
	return nil
}

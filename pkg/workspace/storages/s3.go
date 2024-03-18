package storages

import (
	"bytes"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

// S3Storage is an implementation of workspace.Storage which uses s3 as storage.
type S3Storage struct {
	s3     *s3.S3
	bucket string

	// The prefix to store the workspaces files.
	prefix string
}

// NewS3Storage news s3 workspace storage and init default workspace.
func NewS3Storage(s3 *s3.S3, bucket, prefix string) (*S3Storage, error) {
	s := &S3Storage{
		s3:     s3,
		bucket: bucket,
		prefix: prefix,
	}
	return s, s.initDefaultWorkspaceIf()
}

func (s *S3Storage) Get(name string) (*v1.Workspace, error) {
	exist, err := s.Exist(name)
	if err != nil {
		return nil, err
	}
	if !exist {
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

func (s *S3Storage) Update(ws *v1.Workspace) error {
	exist, err := s.Exist(ws.Name)
	if err != nil {
		return err
	}
	if !exist {
		return ErrWorkspaceNotExist
	}

	return s.writeWorkspace(ws)
}

func (s *S3Storage) Delete(name string) error {
	meta, err := s.readMeta()
	if err != nil {
		return err
	}
	if !checkWorkspaceExistence(meta, name) {
		return nil
	}

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.prefix + "/" + name + yamlSuffix),
	}
	if _, err = s.s3.DeleteObject(input); err != nil {
		return fmt.Errorf("remove workspace in s3 failed: %w", err)
	}

	removeAvailableWorkspaces(meta, name)
	return s.writeMeta(meta)
}

func (s *S3Storage) Exist(name string) (bool, error) {
	meta, err := s.readMeta()
	if err != nil {
		return false, err
	}
	return checkWorkspaceExistence(meta, name), nil
}

func (s *S3Storage) GetNames() ([]string, error) {
	meta, err := s.readMeta()
	if err != nil {
		return nil, err
	}
	return meta.AvailableWorkspaces, nil
}

func (s *S3Storage) GetCurrent() (string, error) {
	meta, err := s.readMeta()
	if err != nil {
		return "", err
	}
	return meta.Current, nil
}

func (s *S3Storage) SetCurrent(name string) error {
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

func (s *S3Storage) initDefaultWorkspaceIf() error {
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

func (s *S3Storage) readMeta() (*workspacesMetaData, error) {
	key := s.prefix + "/" + metadataFile
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    &key,
	}
	output, err := s.s3.GetObject(input)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == s3.ErrCodeNoSuchKey {
			return &workspacesMetaData{}, nil
		}
		return nil, fmt.Errorf("get workspaces meta data from s3 failed: %w", err)
	}
	defer func() {
		_ = output.Body.Close()
	}()

	content, err := io.ReadAll(output.Body)
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

func (s *S3Storage) writeMeta(meta *workspacesMetaData) error {
	content, err := yaml.Marshal(meta)
	if err != nil {
		return fmt.Errorf("yaml marshal workspaces meta data failed: %w", err)
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.prefix + "/" + metadataFile),
		Body:   bytes.NewReader(content),
	}
	if _, err = s.s3.PutObject(input); err != nil {
		return fmt.Errorf("put workspaces meta data to s3 failed: %w", err)
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

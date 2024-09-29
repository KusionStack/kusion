package storages

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/engine/resource/graph"
)

// S3Storage is an implementation of graph.Storage which uses s3 as storage.
type S3Storage struct {
	s3     *s3.S3
	bucket string

	// The prefix to store the graph files.
	prefix string
}

// NewS3Storage news s3 graph storage, and derives metadata.
func NewS3Storage(s3 *s3.S3, bucket, prefix string) (*S3Storage, error) {
	s := &S3Storage{
		s3:     s3,
		bucket: bucket,
		prefix: prefix,
	}
	return s, nil
}

// Get gets the graph from s3.
func (s *S3Storage) Get() (*v1.Graph, error) {
	output, err := getS3StorageObject(s.s3, s.bucket, s.prefix, graphFileName)
	if err != nil {
		return nil, fmt.Errorf("get graph from s3 failed: %w", err)
	}
	defer func() {
		_ = output.Body.Close()
	}()
	content, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("read graph failed: %w", err)
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
func (s *S3Storage) Create(r *v1.Graph) error {
	output, _ := getS3StorageObject(s.s3, s.bucket, s.prefix, graphFileName)
	if output != nil {
		return ErrGraphAlreadyExist
	}

	return s.writeGraph(r)
}

// Update updates the graph in s3.
func (s *S3Storage) Update(r *v1.Graph) error {
	_, err := getS3StorageObject(s.s3, s.bucket, s.prefix, graphFileName)
	if err != nil {
		return ErrGraphNotExist
	}

	return s.writeGraph(r)
}

// Delete deletes the graph in s3
func (s *S3Storage) Delete() error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s.prefix, graphFileName)),
	}
	if _, err := s.s3.DeleteObject(input); err != nil {
		return fmt.Errorf("remove workspace in s3 failed: %w", err)
	}

	return nil
}

// writeGraph writes the graph to s3.
func (s *S3Storage) writeGraph(r *v1.Graph) error {
	content, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("json marshal graph failed: %w", err)
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s.prefix, graphFileName)),
		Body:   bytes.NewReader(content),
	}
	if _, err = s.s3.PutObject(input); err != nil {
		return fmt.Errorf("put graph to s3 failed: %w", err)
	}

	return nil
}

// CheckGraphStorageExistence checks whether the graph storage exists.
func (s *S3Storage) CheckGraphStorageExistence() bool {
	if _, err := getS3StorageObject(s.s3, s.bucket, s.prefix, graphFileName); err != nil {
		return false
	}

	return true
}

// getS3StorageObject gets the graph object from s3.
func getS3StorageObject(s *s3.S3, bucket, prefix, graphFileName string) (*s3.GetObjectOutput, error) {
	key := fmt.Sprintf("%s/%s", prefix, graphFileName)
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    &key,
	}
	output, err := s.GetObject(input)
	if err != nil {
		return nil, fmt.Errorf("get graph from s3 failed: %w", err)
	}

	return output, nil
}

package storages

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/engine/resource/graph"
)

// OssStorage is an implementation of graph.Storage which uses oss as storage.
type OssStorage struct {
	bucket *oss.Bucket

	// The prefix to store the graph files.
	prefix string
}

// NewOssStorage news oss graph storage, and derives metadata.
func NewOssStorage(bucket *oss.Bucket, prefix string) (*OssStorage, error) {
	s := &OssStorage{
		bucket: bucket,
		prefix: prefix,
	}

	return s, nil
}

// Get gets the graph from oss.
func (s *OssStorage) Get() (*v1.Graph, error) {
	body, err := s.bucket.GetObject(fmt.Sprintf("%s/%s", s.prefix, graphFileName))
	if err != nil {
		return nil, fmt.Errorf("get resource graph from oss failed: %w", err)
	}
	defer func() {
		_ = body.Close()
	}()
	content, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read resource graph failed: %w", err)
	}

	r := &v1.Graph{}
	if err = json.Unmarshal(content, r); err != nil {
		return nil, fmt.Errorf("json unmarshal graph failed: %w", err)
	}
	// Index is not stored in oss, so we need to rebuild it.
	// Update resource index to use index in the memory.

	graph.UpdateResourceIndex(r.Resources)

	return r, nil
}

// Create creates the graph in oss.
func (s *OssStorage) Create(r *v1.Graph) error {
	body, _ := s.bucket.GetObject(fmt.Sprintf("%s/%s", s.prefix, graphFileName))
	if body != nil {
		return ErrGraphAlreadyExist
	}
	defer func() {
		_ = body.Close()
	}()

	return s.writeGraph(r)
}

// Update updates the graph in oss.
func (s *OssStorage) Update(r *v1.Graph) error {
	body, err := s.bucket.GetObject(fmt.Sprintf("%s/%s", s.prefix, graphFileName))
	if err != nil {
		return ErrGraphNotExist
	}
	defer func() {
		_ = body.Close()
	}()

	return s.writeGraph(r)
}

// Delete deletes the graph in oss.
func (s *OssStorage) Delete() error {
	if err := s.bucket.DeleteObject(fmt.Sprintf("%s/%s", s.prefix, graphFileName)); err != nil {
		return fmt.Errorf("remove workspace in oss failed: %w", err)
	}

	return nil
}

// writeGraph writes the graph to oss.
func (s *OssStorage) writeGraph(r *v1.Graph) error {
	content, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("json marshal graph failed: %w", err)
	}

	key := fmt.Sprintf("%s/%s", s.prefix, graphFileName)
	if err = s.bucket.PutObject(key, bytes.NewReader(content)); err != nil {
		return fmt.Errorf("put graph to oss failed: %w", err)
	}

	return nil
}

// CheckGraphStorageExistence checks whether the graph storage exists.
func (s *OssStorage) CheckGraphStorageExistence() bool {
	body, err := s.bucket.GetObject(fmt.Sprintf("%s/%s", s.prefix, graphFileName))
	defer func() {
		_ = body.Close()
	}()

	return err == nil
}

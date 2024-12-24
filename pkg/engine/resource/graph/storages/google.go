package storages

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/engine/resource/graph"

	googlestorage "cloud.google.com/go/storage"
)

// GoogleStorage is an implementation of graph.Storage which uses google as storage.
type GoogleStorage struct {
	bucket googlestorage.BucketHandle

	// The prefix to store the release files.
	prefix string
}

// NewGoogleStorage news google cloud graph storage, and derives metadata.
func NewGoogleStorage(bucket *googlestorage.BucketHandle, prefix string) (*GoogleStorage, error) {
	s := &GoogleStorage{
		bucket: *bucket,
		prefix: prefix,
	}
	return s, nil
}

// Get gets the graph from google.
func (s *GoogleStorage) Get() (*v1.Graph, error) {
	output, _ := s.getGoogleStorageObject(s.prefix, graphFileName)
	r := &v1.Graph{}
	if err := json.Unmarshal(output, r); err != nil {
		return nil, fmt.Errorf("json unmarshal graph failed: %w", err)
	}

	// Index is not stored in google, so we need to rebuild it.
	// Update resource index to use index in the memory.
	graph.UpdateResourceIndex(r.Resources)

	return r, nil
}

// Create creates the graph in google.
func (s *GoogleStorage) Create(r *v1.Graph) error {
	output, _ := s.getGoogleStorageObject(s.prefix, graphFileName)
	if output != nil {
		return ErrGraphAlreadyExist
	}

	return s.writeGraph(r)
}

// Update updates the graph in google.
func (s *GoogleStorage) Update(r *v1.Graph) error {
	_, err := s.getGoogleStorageObject(s.prefix, graphFileName)
	if err != nil {
		return ErrGraphNotExist
	}

	return s.writeGraph(r)
}

// Delete deletes the graph in google
func (s *GoogleStorage) Delete() error {
	key := fmt.Sprintf("%s/%s", s.prefix, graphFileName)
	obj := s.bucket.Object(key)
	if err := obj.Delete(context.Background()); err != nil {
		return fmt.Errorf("delete graph from google storage failed: %w", err)
	}

	return nil
}

// writeGraph writes the graph to google.
func (s *GoogleStorage) writeGraph(r *v1.Graph) error {
	content, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("json marshal graph failed: %w", err)
	}

	obj := s.bucket.Object(fmt.Sprintf("%s/%s", s.prefix, graphFileName))
	writer := obj.NewWriter(context.Background())
	if _, err = writer.Write(content); err != nil {
		return fmt.Errorf("write graph failed: %w", err)
	}
	if err = writer.Close(); err != nil {
		return fmt.Errorf("close writer failed: %w", err)
	}
	return nil
}

// CheckGraphStorageExistence checks whether the graph storage exists.
func (s *GoogleStorage) CheckGraphStorageExistence() bool {
	if _, err := s.getGoogleStorageObject(s.prefix, graphFileName); err != nil {
		return false
	}

	return true
}

// getGoogleStorageObject gets the graph object from google.
func (s *GoogleStorage) getGoogleStorageObject(prefix, graphFileName string) ([]byte, error) {
	key := fmt.Sprintf("%s/%s", prefix, graphFileName)
	ctx := context.Background()
	obj := s.bucket.Object(key)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("get release from google storage failed: %w", err)
	}
	defer reader.Close()
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read release failed: %w", err)
	}
	return content, nil
}

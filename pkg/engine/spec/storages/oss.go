package storages

import (
	"bytes"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/engine/spec"
)

// OssStorage should implement the spec.Storage interface.
var _ spec.Storage = &OssStorage{}

type OssStorage struct {
	bucket *oss.Bucket

	// The oss key to store the spec file.
	key string
}

// NewOssStorage constructs an Aliyun OSS based spec storage.
func NewOssStorage(bucket *oss.Bucket, key string) *OssStorage {
	return &OssStorage{
		bucket: bucket,
		key:    key,
	}
}

// Get returns the Spec, if the Spec does not exist, return nil.
func (s *OssStorage) Get() (*v1.Intent, error) {
	var exist bool
	body, err := s.bucket.GetObject(s.key)
	if err != nil {
		ossErr, ok := err.(oss.ServiceError)
		// error code ref: github.com/aliyun/aliyun-oss-go-sdk@v2.1.8+incompatible/oss/bucket.go:553
		if ok && ossErr.StatusCode == 404 {
			exist = true
		}
		if exist {
			return nil, nil
		}
		return nil, err
	}
	defer func() {
		_ = body.Close()
	}()

	content, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	if len(content) == 0 {
		return nil, nil
	}

	intent := &v1.Intent{}
	err = yaml.Unmarshal(content, intent)
	if err != nil {
		return nil, err
	}
	return intent, nil
}

// Apply updates the spec if already exists, or create a new spec.
func (s *OssStorage) Apply(intent *v1.Intent) error {
	content, err := yaml.Marshal(intent)
	if err != nil {
		return err
	}

	return s.bucket.PutObject(s.key, bytes.NewReader(content))
}

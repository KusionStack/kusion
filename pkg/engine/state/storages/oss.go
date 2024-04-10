package storages

import (
	"bytes"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// OssStorage is an implementation of state.Backend which uses oss as storage.
type OssStorage struct {
	bucket *oss.Bucket

	// The oss key to store the state file.
	key string
}

func NewOssStorage(bucket *oss.Bucket, key string) *OssStorage {
	return &OssStorage{
		bucket: bucket,
		key:    key,
	}
}

func (s *OssStorage) Get() (*v1.DeprecatedState, error) {
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

	state := &v1.DeprecatedState{}
	err = yaml.Unmarshal(content, state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (s *OssStorage) Apply(state *v1.DeprecatedState) error {
	content, err := yaml.Marshal(state)
	if err != nil {
		return err
	}

	return s.bucket.PutObject(s.key, bytes.NewReader(content))
}

package oss

import (
	"errors"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/zclconf/go-cty/cty"
	"kusionstack.io/kusion/pkg/engine/states"
)

type OssBackend struct {
	OssState
}

func NewOssBackend() states.Backend {
	return &OssBackend{}
}

// ConfigSchema returns a description of the expected configuration
// structure for the receiving backend.
func (b *OssBackend) ConfigSchema() cty.Type {
	config := map[string]cty.Type{
		"endpoint":        cty.String,
		"bucket":          cty.String,
		"accessKeyID":     cty.String,
		"accessKeySecret": cty.String,
	}
	return cty.Object(config)
}

// Configure uses the provided configuration to set configuration fields
// within the OssState backend.
func (b *OssBackend) Configure(obj cty.Value) error {
	var endpoint, bucket, accessKeyID, accessKeySecret cty.Value
	if endpoint = obj.GetAttr("endpoint"); endpoint.IsNull() {
		return errors.New("oss endpoint must be configure in backend config")
	}
	if bucket = obj.GetAttr("bucket"); bucket.IsNull() {
		return errors.New("oss bucket must be configure in backend config")
	}
	if accessKeyID = obj.GetAttr("accessKeyID"); accessKeyID.IsNull() {
		return errors.New("oss accessKeyID must be configure in backend config")
	}
	if accessKeySecret = obj.GetAttr("accessKeySecret"); accessKeySecret.IsNull() {
		return errors.New("oss accessKeySecret must be configure in backend config")
	}

	ossClient, err := oss.New(endpoint.AsString(), accessKeyID.AsString(), accessKeySecret.AsString())
	if err != nil {
		return nil
	}
	ossBucket, err := ossClient.Bucket(bucket.AsString())
	if err != nil {
		return err
	}
	b.bucket = ossBucket

	return nil
}

// StateStorage return a StateStorage to manage State stored in oss
func (b *OssBackend) StateStorage() states.StateStorage {
	return &OssState{b.bucket}
}

package storages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/internal.kusion.io/v1"
)

func TestValidateOssConfig(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *v1.BackendOssConfig
	}{
		{
			name:    "valid oss config",
			success: true,
			config: &v1.BackendOssConfig{
				GenericBackendObjectStorageConfig: &v1.GenericBackendObjectStorageConfig{
					Endpoint:        "http://oss-cn-hangzhou.aliyuncs.com",
					AccessKeyID:     "fake-access-key-id",
					AccessKeySecret: "fake-access-key-secret",
					Bucket:          "kusion",
				},
			},
		},
		{
			name:    "invalid oss config empty endpoint",
			success: false,
			config: &v1.BackendOssConfig{
				GenericBackendObjectStorageConfig: &v1.GenericBackendObjectStorageConfig{
					Endpoint:        "",
					AccessKeyID:     "fake-access-key-id",
					AccessKeySecret: "fake-access-key-secret",
					Bucket:          "kusion",
				},
			},
		},
		{
			name:    "invalid oss config empty access key id",
			success: false,
			config: &v1.BackendOssConfig{
				GenericBackendObjectStorageConfig: &v1.GenericBackendObjectStorageConfig{
					Endpoint:        "http://oss-cn-hangzhou.aliyuncs.com",
					AccessKeyID:     "",
					AccessKeySecret: "fake-access-key-secret",
					Bucket:          "kusion",
				},
			},
		},
		{
			name:    "invalid oss config empty access key secret",
			success: false,
			config: &v1.BackendOssConfig{
				GenericBackendObjectStorageConfig: &v1.GenericBackendObjectStorageConfig{
					Endpoint:        "http://oss-cn-hangzhou.aliyuncs.com",
					AccessKeyID:     "fake-access-key-id",
					AccessKeySecret: "",
					Bucket:          "kusion",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateOssConfig(tc.config)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateOssConfigFromFile(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *v1.BackendOssConfig
	}{
		{
			name:    "valid oss config from file",
			success: true,
			config: &v1.BackendOssConfig{
				GenericBackendObjectStorageConfig: &v1.GenericBackendObjectStorageConfig{
					Endpoint: "http://oss-cn-hangzhou.aliyuncs.com",
					Bucket:   "kusion",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateOssConfigFromFile(tc.config)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateS3Config(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *v1.BackendS3Config
	}{
		{
			name:    "valid s3 config",
			success: true,
			config: &v1.BackendS3Config{
				GenericBackendObjectStorageConfig: &v1.GenericBackendObjectStorageConfig{
					AccessKeyID:     "fake-access-key-id",
					AccessKeySecret: "fake-access-key-secret",
					Bucket:          "kusion",
				},
				Region: "us-east-1",
			},
		},
		{
			name:    "invalid s3 config empty region",
			success: false,
			config: &v1.BackendS3Config{
				GenericBackendObjectStorageConfig: &v1.GenericBackendObjectStorageConfig{
					AccessKeyID:     "fake-access-key-id",
					AccessKeySecret: "fake-access-key-secret",
					Bucket:          "kusion",
				},
				Region: "",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateS3Config(tc.config)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateS3ConfigFromFile(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *v1.BackendS3Config
	}{
		{
			name:    "valid s3 config from file",
			success: true,
			config: &v1.BackendS3Config{
				GenericBackendObjectStorageConfig: &v1.GenericBackendObjectStorageConfig{
					Bucket: "kusion",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateS3ConfigFromFile(tc.config)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

package storages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

func TestValidateMysqlConfig(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *v1.BackendMysqlConfig
	}{
		{
			name:    "valid mysql config",
			success: true,
			config: &v1.BackendMysqlConfig{
				DBName: "kusion",
				User:   "kk",
				Host:   "127.0.0.1",
				Port:   3306,
			},
		},
		{
			name:    "invalid mysql config empty port",
			success: false,
			config: &v1.BackendMysqlConfig{
				DBName: "kusion",
				User:   "kk",
				Host:   "127.0.0.1",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMysqlConfig(tc.config)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateMysqlConfigFromFile(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *v1.BackendMysqlConfig
	}{
		{
			name:    "valid mysql config",
			success: true,
			config: &v1.BackendMysqlConfig{
				DBName: "kusion",
				User:   "kk",
				Host:   "127.0.0.1",
			},
		},
		{
			name:    "invalid mysql config empty dbName",
			success: false,
			config: &v1.BackendMysqlConfig{
				DBName: "",
				User:   "kk",
				Host:   "127.0.0.1",
			},
		},
		{
			name:    "invalid mysql config empty user",
			success: false,
			config: &v1.BackendMysqlConfig{
				DBName: "kusion",
				User:   "",
				Host:   "127.0.0.1",
			},
		},
		{
			name:    "invalid mysql config empty host",
			success: false,
			config: &v1.BackendMysqlConfig{
				DBName: "kusion",
				User:   "kk",
				Host:   "",
			},
		},
		{
			name:    "invalid mysql config invalid port",
			success: false,
			config: &v1.BackendMysqlConfig{
				DBName: "kusion",
				User:   "kk",
				Host:   "127.0.0.1",
				Port:   -1,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMysqlConfigFromFile(tc.config)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

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
					Bucket:          "kusion_bucket",
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
					Bucket:          "kusion_bucket",
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
					Bucket:          "kusion_bucket",
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
					Bucket:          "kusion_bucket",
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
					Bucket:   "kusion_bucket",
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
					Bucket:          "kusion_bucket",
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
					Bucket:          "kusion_bucket",
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
					Bucket: "kusion_bucket",
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

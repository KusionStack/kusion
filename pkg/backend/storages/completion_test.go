package storages

import (
	"os"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/internal.kusion.io/v1"
	"kusionstack.io/kusion/pkg/util/kfile"
)

func TestCompleteLocalConfig(t *testing.T) {
	testcases := []struct {
		name                 string
		success              bool
		config               *v1.BackendLocalConfig
		mockKusionDataFolder string
		completeConfig       *v1.BackendLocalConfig
	}{
		{
			name:                 "complete local config",
			success:              true,
			config:               &v1.BackendLocalConfig{},
			mockKusionDataFolder: "/etc",
			completeConfig: &v1.BackendLocalConfig{
				Path: "/etc",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock kusion data folder", t, func() {
				mockey.Mock(kfile.KusionDataFolder).Return(tc.mockKusionDataFolder, nil).Build()
				err := CompleteLocalConfig(tc.config)
				assert.Equal(t, tc.success, err == nil)
				assert.Equal(t, tc.completeConfig, tc.config)
			})
		})
	}
}

func TestCompleteOssConfig(t *testing.T) {
	testcases := []struct {
		name           string
		config         *v1.BackendOssConfig
		envs           map[string]string
		completeConfig *v1.BackendOssConfig
	}{
		{
			name: "complete oss config",
			config: &v1.BackendOssConfig{
				GenericBackendObjectStorageConfig: &v1.GenericBackendObjectStorageConfig{
					Endpoint: "http://oss-cn-hangzhou.aliyuncs.com",
					Bucket:   "kusion",
				},
			},
			envs: map[string]string{
				v1.EnvOssAccessKeyID:     "fake-ak",
				v1.EnvOssAccessKeySecret: "fake-sk",
			},
			completeConfig: &v1.BackendOssConfig{
				GenericBackendObjectStorageConfig: &v1.GenericBackendObjectStorageConfig{
					Endpoint:        "http://oss-cn-hangzhou.aliyuncs.com",
					Bucket:          "kusion",
					AccessKeyID:     "fake-ak",
					AccessKeySecret: "fake-sk",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envs {
				_ = os.Setenv(k, v)
			}
			CompleteOssConfig(tc.config)
			assert.Equal(t, tc.completeConfig, tc.config)
			for k := range tc.envs {
				_ = os.Unsetenv(k)
			}
		})
	}
}

func TestCompleteS3Config(t *testing.T) {
	testcases := []struct {
		name           string
		config         *v1.BackendS3Config
		envs           map[string]string
		completeConfig *v1.BackendS3Config
	}{
		{
			name: "complete s3 config",
			config: &v1.BackendS3Config{
				GenericBackendObjectStorageConfig: &v1.GenericBackendObjectStorageConfig{
					Endpoint: "fake-endpoint",
					Bucket:   "kusion",
				},
			},
			envs: map[string]string{
				v1.EnvAwsRegion:          "us-east-1",
				v1.EnvAwsAccessKeyID:     "fake-ak",
				v1.EnvAwsSecretAccessKey: "fake-sk",
			},
			completeConfig: &v1.BackendS3Config{
				GenericBackendObjectStorageConfig: &v1.GenericBackendObjectStorageConfig{
					Endpoint:        "fake-endpoint",
					Bucket:          "kusion",
					AccessKeyID:     "fake-ak",
					AccessKeySecret: "fake-sk",
				},
				Region: "us-east-1",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envs {
				_ = os.Setenv(k, v)
			}
			CompleteS3Config(tc.config)
			assert.Equal(t, tc.completeConfig, tc.config)
			for k := range tc.envs {
				_ = os.Unsetenv(k)
			}
		})
	}
}

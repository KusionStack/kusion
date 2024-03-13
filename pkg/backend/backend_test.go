package backend

import (
	"os"
	"reflect"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/backend/storages"
	"kusionstack.io/kusion/pkg/config"
)

func mockConfig() *v1.Config {
	return &v1.Config{
		Backends: &v1.BackendConfigs{
			Current: "pre",
			Backends: map[string]*v1.BackendConfig{
				"dev": {
					Type: v1.BackendTypeLocal,
					Configs: map[string]any{
						v1.BackendLocalPath: "/etc",
					},
				},
				"pre": {
					Type: v1.BackendTypeMysql,
					Configs: map[string]any{
						v1.BackendMysqlDBName: "kusion",
						v1.BackendMysqlUser:   "kk",
						v1.BackendMysqlHost:   "127.0.0.1",
						v1.BackendMysqlPort:   3306,
					},
				},
				"staging": {
					Type: v1.BackendTypeOss,
					Configs: map[string]any{
						v1.BackendGenericOssEndpoint: "http://oss-cn-hangzhou.aliyuncs.com",
						v1.BackendGenericOssBucket:   "kusion",
					},
				},
				"prod": {
					Type: v1.BackendTypeS3,
					Configs: map[string]any{
						v1.BackendGenericOssBucket: "kusion",
					},
				},
			},
		},
	}
}

func mockCompleteLocalStorage() {
	mockey.Mock(storages.CompleteLocalConfig).Return(nil).Build()
}

func mockNewStorage() {
	mockey.Mock(storages.NewLocalStorage).Return(&storages.LocalStorage{}).Build()
	mockey.Mock(storages.NewMysqlStorage).Return(&storages.MysqlStorage{}, nil).Build()
	mockey.Mock(storages.NewOssStorage).Return(&storages.OssStorage{}, nil).Build()
	mockey.Mock(storages.NewS3Storage).Return(&storages.S3Storage{}, nil).Build()
}

func TestNewBackend(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		cfg     *v1.Config
		envs    map[string]string
		bkName  string
		storage Backend
	}{
		{
			name:    "new default backend",
			success: true,
			cfg: func() *v1.Config {
				cfg := mockConfig()
				cfg.Backends.Current = ""
				return cfg
			}(),
			envs:    nil,
			bkName:  "",
			storage: &storages.LocalStorage{},
		},
		{
			name:    "new current backend",
			success: true,
			cfg:     mockConfig(),
			envs:    nil,
			bkName:  "",
			storage: &storages.MysqlStorage{},
		},
		{
			name:    "new local backend",
			success: true,
			cfg:     mockConfig(),
			envs:    nil,
			bkName:  "dev",
			storage: &storages.LocalStorage{},
		},
		{
			name:    "new local backend",
			success: true,
			cfg:     mockConfig(),
			envs:    nil,
			bkName:  "dev",
			storage: &storages.LocalStorage{},
		},
		{
			name:    "new mysql backend",
			success: true,
			cfg:     mockConfig(),
			envs: map[string]string{
				v1.EnvBackendMysqlPassword: "fake-password",
			},
			bkName:  "pre",
			storage: &storages.MysqlStorage{},
		},
		{
			name:    "new oss backend",
			success: true,
			cfg:     mockConfig(),
			envs: map[string]string{
				v1.EnvOssAccessKeyID:     "fake-ak",
				v1.EnvOssAccessKeySecret: "fake-sk",
			},
			bkName:  "staging",
			storage: &storages.OssStorage{},
		},
		{
			name:    "new s3 backend",
			success: true,
			cfg:     mockConfig(),
			envs: map[string]string{
				v1.EnvAwsRegion:          "us-east-1",
				v1.EnvAwsAccessKeyID:     "fake-ak",
				v1.EnvAwsSecretAccessKey: "fake-sk",
			},
			bkName:  "prod",
			storage: &storages.S3Storage{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock config", t, func() {
				mockey.Mock(config.GetConfig).Return(tc.cfg, nil).Build()
				mockCompleteLocalStorage()
				mockNewStorage()
				for k, v := range tc.envs {
					_ = os.Setenv(k, v)
				}
				storage, err := NewBackend(tc.bkName)
				assert.Equal(t, tc.success, err == nil)
				assert.Equal(t, reflect.TypeOf(tc.storage), reflect.TypeOf(storage))
				for k := range tc.envs {
					_ = os.Unsetenv(k)
				}
			})
		})
	}
}

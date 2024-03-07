package storages

import (
	"os"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

// CompleteMysqlConfig sets default value of mysql config if not set.
func CompleteMysqlConfig(config *v1.BackendMysqlConfig) {
	if config.Port == 0 {
		config.Port = v1.DefaultMysqlPort
	}
	password := os.Getenv(v1.EnvBackendMysqlPassword)
	if password != "" {
		config.Password = password
	}
}

// CompleteOssConfig constructs the whole oss config by environment variables if set.
func CompleteOssConfig(config *v1.BackendOssConfig) {
	accessKeyID := os.Getenv(v1.EnvOssAccessKeyID)
	accessKeySecret := os.Getenv(v1.EnvOssAccessKeySecret)

	if accessKeyID != "" {
		config.AccessKeyID = accessKeyID
	}
	if accessKeySecret != "" {
		config.AccessKeySecret = accessKeySecret
	}
}

// CompleteS3Config constructs the whole s3 config by environment variables if set.
func CompleteS3Config(config *v1.BackendS3Config) {
	accessKeyID := os.Getenv(v1.EnvAwsAccessKeyID)
	accessKeySecret := os.Getenv(v1.EnvAwsSecretAccessKey)
	region := os.Getenv(v1.EnvAwsRegion)
	if region == "" {
		region = os.Getenv(v1.EnvAwsDefaultRegion)
	}

	if accessKeyID != "" {
		config.AccessKeyID = accessKeyID
	}
	if accessKeySecret != "" {
		config.AccessKeySecret = accessKeySecret
	}
	if region != "" {
		config.Region = region
	}
}

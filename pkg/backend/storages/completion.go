package storages

import (
	"os"

	v1 "kusionstack.io/kusion/pkg/apis/internal.kusion.io/v1"
	"kusionstack.io/kusion/pkg/clipath"
)

// CompleteLocalConfig sets default value of path if not set, which uses the path of kusion data folder.
func CompleteLocalConfig(config *v1.BackendLocalConfig) error {
	if config.Path == "" {
		path, err := clipath.DataPath()
		if err != nil {
			return err
		}
		config.Path = path
	}
	return nil
}

// CompleteMysqlConfig sets default value of port if not set, which is 3306, and fulfills password from environment
// variables if set.
func CompleteMysqlConfig(config *v1.BackendMysqlConfig) {
	if config.Port == 0 {
		config.Port = v1.DefaultMysqlPort
	}
	password := os.Getenv(v1.EnvBackendMysqlPassword)
	if password != "" {
		config.Password = password
	}
}

// CompleteOssConfig fulfills the whole oss config from environment variables if set.
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

// CompleteS3Config fulfills the whole s3 config from environment variables if set.
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

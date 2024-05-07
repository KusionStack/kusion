package storages

import (
	"os"

	v1 "kusionstack.io/kusion/pkg/apis/internal.kusion.io/v1"
	"kusionstack.io/kusion/pkg/util/kfile"
)

// CompleteLocalConfig sets default value of path if not set, which uses the path of kusion data folder.
func CompleteLocalConfig(config *v1.BackendLocalConfig) error {
	if config.Path == "" {
		path, err := kfile.KusionDataFolder()
		if err != nil {
			return err
		}
		config.Path = path
	}
	return nil
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

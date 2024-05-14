package storages

import (
	"errors"
	"fmt"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

var (
	ErrEmptyBucket          = errors.New("empty bucket")
	ErrEmptyAccessKeyID     = errors.New("empty access key id")
	ErrEmptyAccessKeySecret = errors.New("empty access key secret")
	ErrEmptyOssEndpoint     = errors.New("empty oss endpoint")
	ErrEmptyS3Region        = errors.New("empty s3 region")
)

// ValidateOssConfig is used to validate v1.BackendOssConfig is valid or not, where all the items are included.
// If valid, the config contains all valid items to new an oss client.
func ValidateOssConfig(config *v1.BackendOssConfig) error {
	if err := ValidateOssConfigFromFile(config); err != nil {
		return err
	}
	if err := validateGenericObjectStorageSecret(config.AccessKeyID, config.AccessKeySecret); err != nil {
		return fmt.Errorf("%w of %s", err, v1.BackendTypeOss)
	}
	return nil
}

// ValidateOssConfigFromFile is used to validate the v1.BackendOssConfig parsed from config file is valid or not,
// where the sensitive data items set as environment variables are not included.
func ValidateOssConfigFromFile(config *v1.BackendOssConfig) error {
	if err := validateGenericObjectStorageBucket(config.Bucket); err != nil {
		return fmt.Errorf("%w of %s", err, v1.BackendTypeOss)
	}
	if config.Endpoint == "" {
		return ErrEmptyOssEndpoint
	}
	return nil
}

// ValidateS3Config is used to validate s3Config is valid or not, where all the items are included.
// If valid, the config  contains all valid items to new a s3 client.
func ValidateS3Config(config *v1.BackendS3Config) error {
	if err := ValidateS3ConfigFromFile(config); err != nil {
		return err
	}
	if err := validateGenericObjectStorageSecret(config.AccessKeyID, config.AccessKeySecret); err != nil {
		return fmt.Errorf("%w of %s", err, v1.BackendTypeS3)
	}
	if config.Region == "" {
		return ErrEmptyS3Region
	}
	return nil
}

// ValidateS3ConfigFromFile is used to validate the v1.BackendS3Config parsed from config file is valid or not,
// where the sensitive data items set as environment variables are not included.
func ValidateS3ConfigFromFile(config *v1.BackendS3Config) error {
	if err := validateGenericObjectStorageBucket(config.Bucket); err != nil {
		return fmt.Errorf("%w of %s", err, v1.BackendTypeS3)
	}
	return nil
}

func validateGenericObjectStorageBucket(bucket string) error {
	if bucket == "" {
		return ErrEmptyBucket
	}
	return nil
}

func validateGenericObjectStorageSecret(ak, sk string) error {
	if ak == "" {
		return ErrEmptyAccessKeyID
	}
	if sk == "" {
		return ErrEmptyAccessKeySecret
	}
	return nil
}

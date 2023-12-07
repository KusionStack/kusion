package fake

import (
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/models"
)

type (
	GetSecretInfoFn      func(secretName string) (*models.SecretInfo, error)
	SecretsManagerClient struct {
		GetSecretInfoFn GetSecretInfoFn
	}
)

func NewGetSecretInfoFn(secretValue interface{}, secretDataType string, err error) GetSecretInfoFn {
	return func(secretName string) (*models.SecretInfo, error) {
		if secretValue == nil {
			return nil, err
		}
		if secretDataType == "text" {
			secretString := secretValue.(string)
			return &models.SecretInfo{
				SecretValue: secretString,
			}, err
		}
		if secretDataType == "binary" {
			secretBinary := secretValue.([]byte)
			return &models.SecretInfo{
				SecretValueByteBuffer: secretBinary,
			}, err
		}

		return nil, err
	}
}

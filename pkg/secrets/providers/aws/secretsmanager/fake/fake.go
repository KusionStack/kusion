package fake

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

type (
	GetSecretValueFn     func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
	SecretsManagerClient struct {
		GetSecretValueFn GetSecretValueFn
	}
)

func NewGetSecretValueFn(secretData interface{}, dataType string, err error) GetSecretValueFn {
	return func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
		if secretData == nil {
			return nil, &types.ResourceNotFoundException{}
		}
		if dataType == "string" {
			secretString := secretData.(string)
			return &secretsmanager.GetSecretValueOutput{
				SecretString: &secretString,
			}, err
		}
		secretBinary := secretData.([]byte)
		return &secretsmanager.GetSecretValueOutput{
			SecretBinary: secretBinary,
		}, err
	}
}

func (sc *SecretsManagerClient) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return sc.GetSecretValueFn(ctx, params, optFns...)
}

package fake

import (
	"context"

	"kusionstack.io/kusion/pkg/secrets/providers/viettelcloud/vclient"
)

type (
	SecretManagerSecretsRetrieveWithResponseFn func(ctx context.Context, name string, params *vclient.SecretManagerSecretsRetrieveParams, reqEditors ...vclient.RequestEditorFn) (*vclient.SecretManagerSecretsRetrieveResponse, error)
	SecretsManagerClient                       struct {
		SecretManagerSecretsRetrieveWithResponseFn SecretManagerSecretsRetrieveWithResponseFn
	}
)

func NewSecretManagerSecretsRetrieveWithResponseFn(secretValue map[string]interface{}, secretDataType string, err error) SecretManagerSecretsRetrieveWithResponseFn {
	return func(ctx context.Context, name string, params *vclient.SecretManagerSecretsRetrieveParams, reqEditors ...vclient.RequestEditorFn) (*vclient.SecretManagerSecretsRetrieveResponse, error) {
		if secretValue == nil {
			return nil, err
		}
		if secretDataType == "key-value" {
			return &vclient.SecretManagerSecretsRetrieveResponse{
				JSON200: &vclient.SecretRetrieve{
					Secret: &secretValue,
				},
			}, err
		}
		return nil, err
	}
}

func (sc *SecretsManagerClient) SecretManagerSecretsRetrieveWithResponse(ctx context.Context, name string, params *vclient.SecretManagerSecretsRetrieveParams, reqEditors ...vclient.RequestEditorFn) (*vclient.SecretManagerSecretsRetrieveResponse, error) {
	return sc.SecretManagerSecretsRetrieveWithResponseFn(ctx, name, params, reqEditors...)
}

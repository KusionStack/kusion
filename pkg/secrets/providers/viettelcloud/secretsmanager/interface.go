package secretsmanager

import (
	"context"

	"kusionstack.io/kusion/pkg/secrets/providers/viettelcloud/vclient"
)

// Client is a testable interface for making operations call for AWS Secrets Manager.
type Client interface {
	SecretManagerSecretsRetrieveWithResponse(ctx context.Context, name string, params *vclient.SecretManagerSecretsRetrieveParams, reqEditors ...vclient.RequestEditorFn) (*vclient.SecretManagerSecretsRetrieveResponse, error)
}

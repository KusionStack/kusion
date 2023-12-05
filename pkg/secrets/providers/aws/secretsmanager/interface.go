package secretsmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// Client is a testable interface for making operations call for AWS Secrets Manager.
type Client interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

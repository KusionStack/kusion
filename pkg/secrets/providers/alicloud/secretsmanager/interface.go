package secretsmanager

import (
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/models"
)

// Client is a testable interface for making operations call for Alicloud Secrets Manager.
type Client interface {
	GetSecretInfo(secretName string) (*models.SecretInfo, error)
}

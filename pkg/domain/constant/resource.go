package constant

import "errors"

const (
	AWSProviderType         = "aws"
	AliCloudProviderType    = "alicloud"
	AzureProviderType       = "azure"
	GoogleProviderType      = "google"
	CustomProviderType      = "custom"
	HashicorpProviderType   = "hashicorp"
	StatusResourceApplied   = "applied"
	StatusResourceDestroyed = "destroyed"
	StatusResourceFailed    = "failed"
	StatusResourceUnknown   = "unknown"
	TmpDirPrefix            = "/tmp"
)

var ErrResourceHasNilStack = errors.New("resource has nil stack")

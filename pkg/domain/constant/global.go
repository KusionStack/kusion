package constant

import (
	"errors"
	"time"
)

// These constants represent the possible states of a stack.
const (
	DefaultUser             = "test.user"
	DefaultWorkspace        = "default"
	DefaultBackend          = "default"
	DefaultOrgOwner         = "kusion"
	DefaultSourceType       = SourceProviderTypeGit
	DefaultSourceDesc       = "Default source"
	DefaultSystemName       = "kusion"
	DefaultReleaseNamespace = "server"
	MaxConcurrent           = 10
	MaxAsyncConcurrent      = 1
	MaxAsyncBuffer          = 100
	DefaultLogFilePath      = "/home/admin/logs/kusion.log"
	RepoCacheTTL            = 60 * time.Minute
	RunTimeOut              = 60 * time.Minute
	DefaultWorkloadSig      = "kusion.io/is-workload"
	ResourcePageDefault     = 1
	ResourcePageSizeDefault = 100
	ResourcePageSizeLarge   = 1000
	CommonPageDefault       = 1
	CommonPageSizeDefault   = 10
	SortByCreateTimestamp   = "createTimestamp"
	SortByModifiedTimestamp = "modifiedTimestamp"
	SortByName              = "name"
	SortByID                = "id"
)

var (
	ErrEmptyURL   = errors.New("URL is empty")
	ErrInvalidURL = errors.New("invalid URL")
)

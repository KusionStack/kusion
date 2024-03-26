package entity

import (
	"context"

	"kusionstack.io/kusion/pkg/domain/constant"
)

// The SourceProvider represents the abstraction of the source provider(s)
// management framework.
type SourceProvider interface {
	// Get the type of the source provider.
	Type() constant.SourceProviderType
	// Get source and return directory.
	Get(ctx context.Context, opts ...GetOption) (string, error)
	// Cleanup is invoked to cleanup temp resources for the source.
	Cleanup(ctx context.Context)
}

type GetConfig struct {
	Paths []string
	Type  *constant.SourceProviderType
}

type GetOption func(opt *GetConfig)

func WithPaths(paths ...string) GetOption {
	return func(opt *GetConfig) {
		opt.Paths = paths
	}
}

func WithType(typ constant.SourceProviderType) GetOption {
	return func(opt *GetConfig) {
		opt.Type = &typ
	}
}

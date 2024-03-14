package client

import (
	"context"
	"fmt"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

var (
	// CanonicalConfigMediaType is the OCI media type for the config layer.
	CanonicalConfigMediaType types.MediaType = "application/vnd.io.kusionstack.config.v1+json"

	// CanonicalMediaTypePrefix is the suffix for OCI media type for the content layer.
	CanonicalMediaTypePrefix types.MediaType = "application/vnd.io.kusionstack.content.v1"

	// CanonicalContentMediaType is the OCI media type for the content layer.
	CanonicalContentMediaType = types.MediaType(fmt.Sprintf("%s.tar+gzip", CanonicalMediaTypePrefix))
)

// ClientOptions are options for configuring the client behavior.
type ClientOptions struct {
	craneOptions []crane.Option
}

// ClientOption is a function for configuring ClientOptions.
type ClientOption func(o *ClientOptions)

// WithUserAgent sets User-Agent header for any HTTP requests.
func WithUserAgent(userAgent string) ClientOption {
	return func(o *ClientOptions) {
		o.craneOptions = append(o.craneOptions, crane.WithUserAgent(userAgent))
	}
}

// Client provides methods to interact with OCI registry.
type Client struct {
	opts *ClientOptions
}

// NewClient returns a client instance for use communicating with OCI registry.
func NewClient(options ...ClientOption) *Client {
	client := &Client{
		opts: &ClientOptions{},
	}

	// Apply all options
	for _, opt := range options {
		opt(client.opts)
	}

	return client
}

// optionsWithContext returns the crane options for the given context.
func (c *Client) optionsWithContext(ctx context.Context) []crane.Option {
	options := []crane.Option{
		crane.WithContext(ctx),
	}
	return append(options, c.opts.craneOptions...)
}

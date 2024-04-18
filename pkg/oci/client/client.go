package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

var (
	// CanonicalConfigMediaType is the OCI media type for the config layer.
	CanonicalConfigMediaType types.MediaType = "application/vnd.io.kusion.config.v1+json"

	// CanonicalMediaTypePrefix is the suffix for OCI media type for the content layer.
	CanonicalMediaTypePrefix types.MediaType = "application/vnd.io.kusion.content.v1"

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

// WithCredentials sets authenticator for remote operations.
func WithCredentials(credentials string) ClientOption {
	return func(o *ClientOptions) {
		if len(credentials) == 0 {
			return
		}

		var authConfig authn.AuthConfig
		parts := strings.SplitN(credentials, ":", 2)

		if len(parts) == 1 {
			authConfig = authn.AuthConfig{RegistryToken: parts[0]}
		} else {
			authConfig = authn.AuthConfig{Username: parts[0], Password: parts[1]}
		}
		o.craneOptions = append(o.craneOptions, crane.WithAuth(authn.FromConfig(authConfig)))
	}
}

// WithInsecure returns a ClientOption which allows image references to be fetched without TLS.
func WithInsecure(insecure bool) ClientOption {
	return func(o *ClientOptions) {
		if insecure {
			o.craneOptions = append(o.craneOptions, crane.Insecure)
		}
	}
}

// WithPlatform sets a platform for the client.
func WithPlatform(platform *v1.Platform) ClientOption {
	return func(o *ClientOptions) {
		o.craneOptions = append(o.craneOptions, crane.WithPlatform(platform))
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

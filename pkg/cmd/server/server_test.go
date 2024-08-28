package server

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/server"
)

func TestDefaultBackendOptions_ApplyTo(t *testing.T) {
	config := &server.Config{}
	options := DefaultBackendOptions{
		BackendName:     "backend",
		BackendType:     "local",
		BackendEndpoint: "http://example.com",
	}

	err := options.ApplyTo(config)

	assert.NoError(t, err)
	assert.Equal(t, "local", config.DefaultBackend.BackendConfig.Type)
	assert.Equal(t, "backend", config.DefaultBackend.BackendConfig.Configs["bucket"])
	assert.Equal(t, "http://example.com", config.DefaultBackend.BackendConfig.Configs["endpoint"])
	assert.Equal(t, os.Getenv("BACKEND_ACCESS_KEY_ID"), config.DefaultBackend.BackendConfig.Configs["accessKeyID"])
	assert.Equal(t, os.Getenv("BACKEND_ACCESS_KEY_SECRET"), config.DefaultBackend.BackendConfig.Configs["accessKeySecret"])
}

func TestDefaultSourceOptions_ApplyTo(t *testing.T) {
	config := &server.Config{}
	options := &DefaultSourceOptions{
		SourceRemote: "https://github.com/mockorg/mockrepo",
	}

	err := options.ApplyTo(config)
	require.NoError(t, err)

	expectedRemote, _ := url.Parse("https://github.com/mockorg/mockrepo")
	expectedProvider := constant.DefaultSourceType
	expectedDescription := constant.DefaultSourceDesc

	require.Equal(t, expectedRemote, config.DefaultSource.Remote)
	require.Equal(t, expectedProvider, config.DefaultSource.SourceProvider)
	require.Equal(t, expectedDescription, config.DefaultSource.Description)
}

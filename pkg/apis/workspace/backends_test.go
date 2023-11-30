package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockValidBackendConfigs() BackendConfigs {
	return BackendConfigs{
		Local: mockValidLocalBackendConfig(),
	}
}

func mockValidLocalBackendConfig() LocalFileConfig {
	return LocalFileConfig{
		Path: "/etc/.kusion",
	}
}

func TestBackendConfigs_Validate(t *testing.T) {
	testcases := []struct {
		name           string
		success        bool
		backendConfigs BackendConfigs
	}{
		{
			name:           "valid backend configs",
			success:        true,
			backendConfigs: mockValidBackendConfigs(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.backendConfigs.Validate()
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestLocalBackendConfig_Validate(t *testing.T) {
	testcases := []struct {
		name               string
		success            bool
		localBackendConfig LocalFileConfig
	}{
		{
			name:               "valid local backend config",
			success:            true,
			localBackendConfig: mockValidLocalBackendConfig(),
		},
		{
			name:               "invalid local backend config empty path",
			success:            false,
			localBackendConfig: LocalFileConfig{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.localBackendConfig.Validate()
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

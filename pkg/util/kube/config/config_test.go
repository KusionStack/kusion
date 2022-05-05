package config

import (
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func mockGetenv(result string) {
	monkey.Patch(os.Getenv, func(key string) string {
		return result
	})
}

func TestGetKubeConfig(t *testing.T) {
	// Mock
	defer monkey.UnpatchAll()
	mockGetenv("")
	assert.Equal(t, RecommendedKubeConfigFile, GetKubeConfig())
	mockGetenv("test")
	assert.Equal(t, "test", GetKubeConfig())
}

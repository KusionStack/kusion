//go:build !arm64
// +build !arm64

package config

import (
	"github.com/bytedance/mockey"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockGetenv(result string) {
	mockey.Mock(os.Getenv).To(func(key string) string {
		return result
	}).Build()
}

func TestGetKubeConfig(t *testing.T) {
	// Mock
	mockey.PatchConvey("test null env config", t, func() {
		mockGetenv("")
		assert.Equal(t, RecommendedKubeConfigFile, GetKubeConfig())
	})
	mockey.PatchConvey("test env config", t, func() {
		mockGetenv("test")
		assert.Equal(t, "test", GetKubeConfig())
	})
}

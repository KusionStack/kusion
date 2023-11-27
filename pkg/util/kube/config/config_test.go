//go:build !arm64
// +build !arm64

package config

import (
	"os"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/apis/stack"
)

func mockGetenv(result string) {
	mockey.Mock(os.Getenv).To(func(key string) string {
		return result
	}).Build()
}

func TestGetKubeConfig(t *testing.T) {
	stack := &stack.Stack{
		Configuration: stack.Configuration{
			KubeConfig: "",
		},
	}

	// Mock
	mockey.PatchConvey("test null env config", t, func() {
		mockGetenv("")
		assert.Equal(t, RecommendedKubeConfigFile, GetKubeConfig(stack))
	})
	mockey.PatchConvey("test env config", t, func() {
		mockGetenv("test")
		assert.Equal(t, "test", GetKubeConfig(stack))
	})
	mockey.PatchConvey("test stack config", t, func() {
		mockGetenv("")
		stack.KubeConfig = "/home/test/kubeconfig"
		assert.Equal(t, "/home/test/kubeconfig", GetKubeConfig(stack))
	})
}

package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/util/kfile"
)

func testDataFolder() string {
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, "testdata")
}

func mockStack(name string) *v1.Stack {
	return &v1.Stack{
		Name: name,
		Path: fmt.Sprintf("/test_project/%s", name),
	}
}

func Test_NewStateStorage(t *testing.T) {
	testcases := []struct {
		name                     string
		success                  bool
		stack                    *v1.Stack
		opts                     *BackendOptions
		setEnvFunc, unsetEnvFunc func()
	}{
		{
			name:    "default state storage not exist workspace",
			success: true,
			stack:   mockStack("empty_backend_ws_not_exist"),
			opts:    &BackendOptions{},
		},
		{
			name:    "oss state storage use workspace",
			success: true,
			stack:   mockStack("s3_backend_ws"),
			opts:    &BackendOptions{},
			setEnvFunc: func() {
				_ = os.Setenv(v1.EnvAwsRegion, "ua-east-2")
				_ = os.Setenv(v1.EnvAwsAccessKeyID, "aws_ak_id")
				_ = os.Setenv(v1.EnvAwsSecretAccessKey, "aws_ak_secret")
			},
			unsetEnvFunc: func() {
				_ = os.Unsetenv(v1.EnvAwsDefaultRegion)
				_ = os.Unsetenv(v1.EnvOssAccessKeyID)
				_ = os.Unsetenv(v1.EnvAwsSecretAccessKey)
			},
		},
		{
			name:         "invalid workspace",
			success:      false,
			stack:        mockStack("invalid_ws"),
			opts:         &BackendOptions{},
			setEnvFunc:   nil,
			unsetEnvFunc: nil,
		},
		{
			name:         "invalid backend config",
			success:      false,
			stack:        mockStack("invalid_backend_ws"),
			opts:         &BackendOptions{},
			setEnvFunc:   nil,
			unsetEnvFunc: nil,
		},
		{
			name:         "invalid options",
			success:      false,
			stack:        mockStack("not_exist_ws"),
			opts:         &BackendOptions{Type: "not_support"},
			setEnvFunc:   nil,
			unsetEnvFunc: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock kusion data folder", t, func() {
				mockey.Mock(kfile.KusionDataFolder).Return(testDataFolder(), nil).Build()

				if tc.setEnvFunc != nil {
					tc.setEnvFunc()
				}
				_, err := NewStateStorage(tc.stack, tc.opts)
				if tc.unsetEnvFunc != nil {
					tc.unsetEnvFunc()
				}
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

package util

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testFilePath(fileName string) string {
	pwd, _ := os.Getwd()
	return path.Join(pwd, "testdata", fileName)
}

func TestGetValidWorkspaceFromFile(t *testing.T) {
	testcases := []struct {
		name     string
		filePath string
		wsName   string
		success  bool
	}{
		{
			name:     "successfully get workspace",
			filePath: testFilePath("valid_ws.yaml"),
			wsName:   "valid_ws",
			success:  true,
		},
		{
			name:     "failed to get workspace invalid configuration content",
			filePath: testFilePath("invalid_ws.yaml"),
			wsName:   "invalid_ws",
			success:  false,
		},
		{
			name:     "failed to get workspace not exist file",
			filePath: testFilePath("not_exist_ws.yaml"),
			wsName:   "not_exist_ws",
			success:  false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GetValidWorkspaceFromFile(tc.filePath, tc.wsName)
			fmt.Println(err)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

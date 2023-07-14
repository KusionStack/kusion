//go:build !arm64
// +build !arm64

package version

import (
	"encoding/json"
	"errors"
	"reflect"
	"runtime"
	"runtime/debug"
	"testing"
	"time"
	"unsafe"

	"bou.ke/monkey"

	goversion "github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"

	git "kusionstack.io/kusion/pkg/util/gitutil"
)

func TestKusionVersionNormal(t *testing.T) {
	defer monkey.UnpatchAll()
	mockGit()
	mockDependency()
	mockTime()
	mockRuntime()

	versionJSON := `{
		"releaseVersion": "v0.3.11-alpha",
		"gitInfo": {
			"latestTag": "v0.3.11-alpha",
			"commit": "af79cd231e7ed1dbb00e860da9615febf5f17bf0",
			"treeState": "clean"
		},
		"buildInfo": {
			"goVersion": "go1.16.5",
			"GOOS": "` + runtime.GOOS + `",
			"GOARCH": "` + runtime.GOARCH + `",
			"numCPU": 8,
			"compiler": "` + runtime.Compiler + `",
			"buildTime": "2006-01-02 15:04:05"
		},
		"dependency": {
			"kclGoVersion": "stable",
			"kclPluginVersion": "stable"
		}
	}`

	var want Info
	err := json.Unmarshal([]byte(versionJSON), &want)
	if err != nil {
		t.Fatal("unmarshal versionJSON failed", err)
	}
	version, _ := NewInfo()
	assert.Equal(t, want.JSON(), version.JSON())
	assert.Equal(t, want.YAML(), version.YAML())
	assert.Equal(t, want.String(), version.String())
	assert.Equal(t, want.ShortString(), version.ShortString())
}

func TestKusionVersionReturnError(t *testing.T) {
	defer monkey.UnpatchAll()
	mockGit()
	mockDependency()
	mockTime()
	mockRuntime()
	tests := []struct {
		name        string
		specialMock func()
	}{
		{
			name: "head-hash-error",
			specialMock: func() {
				monkey.Patch(git.GetHeadHash, func() (string, error) {
					return "", errors.New("test error")
				})
			},
		},
		{
			name: "head-hash-short-error",
			specialMock: func() {
				monkey.Patch(git.GetHeadHashShort, func() (string, error) {
					return "", errors.New("test error")
				})
			},
		},
		{
			name: "latest-tag-error",
			specialMock: func() {
				monkey.Patch(git.GetLatestTag, func() (string, error) {
					return "", errors.New("test error")
				})
			},
		},
		{
			name: "git-version-error",
			specialMock: func() {
				monkey.Patch(goversion.NewVersion, func(v string) (*goversion.Version, error) {
					return nil, errors.New("test error")
				})
			},
		},
		{
			name: "is-head-at-tag-error",
			specialMock: func() {
				monkey.Patch(git.IsHeadAtTag, func(tag string) (bool, error) {
					return false, errors.New("test error")
				})
			},
		},
		{
			name: "is-dirty-error",
			specialMock: func() {
				monkey.Patch(git.IsDirty, func() (bool, error) {
					return false, errors.New("test error")
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monkey.UnpatchAll()
			mockGit()
			mockDependency()
			mockTime()
			mockRuntime()
			tt.specialMock()
			version, err := NewInfo()
			assert.Nil(t, version)
			assert.Error(t, err)
		},
		)
	}
}

func TestKusionVersionNotHeadTag(t *testing.T) {
	defer monkey.UnpatchAll()
	mockGit()
	mockDependency()
	mockTime()
	mockRuntime()

	monkey.Patch(git.IsHeadAtTag, func(tag string) (bool, error) {
		return false, nil
	})

	versionJSON := `{
	"releaseVersion": "v0.3.11-alpha+af79cd23",
	"gitInfo": {
		"latestTag": "v0.3.11-alpha",
		"commit": "af79cd231e7ed1dbb00e860da9615febf5f17bf0",
		"treeState": "clean"
	},
	"buildInfo": {
		"goVersion": "go1.16.5",
		"GOOS": "` + runtime.GOOS + `",
		"GOARCH": "` + runtime.GOARCH + `",
		"numCPU": 8,
		"compiler": "` + runtime.Compiler + `",
		"buildTime": "2006-01-02 15:04:05"
	},
	"dependency": {
		"kclGoVersion": "stable",
		"kclPluginVersion": "stable"
	}
}`

	var want Info
	err := json.Unmarshal([]byte(versionJSON), &want)
	if err != nil {
		t.Fatal("unmarshal versionJSON failed", err)
	}
	version, _ := NewInfo()
	assert.Equal(t, want.JSON(), version.JSON())
	assert.Equal(t, want.YAML(), version.YAML())
	assert.Equal(t, want.String(), version.String())
	assert.Equal(t, want.ShortString(), version.ShortString())
}

func mockGit() {
	monkey.Patch(git.GetHeadHash, func() (string, error) {
		return "af79cd231e7ed1dbb00e860da9615febf5f17bf0", nil
	})
	monkey.Patch(git.GetHeadHashShort, func() (string, error) {
		return "af79cd23", nil
	})
	monkey.Patch(git.GetLatestTag, func() (string, error) {
		return "v0.3.11-alpha", nil
	})
	monkey.Patch(goversion.NewVersion, func(v string) (*goversion.Version, error) {
		version := &goversion.Version{}
		val := reflect.ValueOf(version).Elem().FieldByName("original")
		reflect.NewAt(val.Type(), unsafe.Pointer(val.UnsafeAddr())).Elem().Set(reflect.ValueOf("v0.3.11-alpha"))
		return version, nil
	})
	monkey.Patch(git.IsHeadAtTag, func(tag string) (bool, error) {
		return true, nil
	})
	monkey.Patch(git.IsDirty, func() (bool, error) {
		return false, nil
	})
}

func mockDependency() {
	monkey.Patch(debug.ReadBuildInfo, func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Deps: []*debug.Module{{Path: KclGoModulePath, Version: "stable"}, {Path: KclPluginModulePath, Version: "stable"}}}, true
	})
}

func mockTime() {
	monkey.Patch(time.Now, func() time.Time {
		t, _ := time.Parse("2006-01-02 15:04:05", "2006-01-02 15:04:05")
		return t
	})
}

func mockRuntime() {
	monkey.Patch(runtime.Version, func() string {
		return "go1.16.5"
	})
	monkey.Patch(runtime.NumCPU, func() int {
		return 8
	})
}

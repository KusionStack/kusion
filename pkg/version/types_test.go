//go:build !arm64
// +build !arm64

package version

import (
	"encoding/json"
	"errors"
	"github.com/bytedance/mockey"
	"reflect"
	"runtime"
	"runtime/debug"
	"testing"
	"time"
	"unsafe"

	goversion "github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"

	git "kusionstack.io/kusion/pkg/util/gitutil"
)

func TestKusionVersionNormal(t *testing.T) {
	mockey.PatchConvey("test kusion version normal", t, func() {
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
	})
}

func TestKusionVersionReturnError(t *testing.T) {
	tests := []struct {
		name        string
		specialMock func()
	}{
		{
			name: "head-hash-error",
			specialMock: func() {
				mockey.Mock(git.GetHeadHash).To(func() (string, error) {
					return "", errors.New("test error")
				}).Build()
			},
		},
		{
			name: "head-hash-short-error",
			specialMock: func() {
				mockey.Mock(git.GetHeadHashShort).To(func() (string, error) {
					return "", errors.New("test error")
				}).Build()
			},
		},
		{
			name: "latest-tag-error",
			specialMock: func() {
				mockey.Mock(git.GetLatestTag).To(func() (string, error) {
					return "", errors.New("test error")
				}).Build()
			},
		},
		{
			name: "git-version-error",
			specialMock: func() {
				mockey.Mock(goversion.NewVersion).To(func(v string) (*goversion.Version, error) {
					return nil, errors.New("test error")
				}).Build()
			},
		},
		{
			name: "is-head-at-tag-error",
			specialMock: func() {
				mockey.Mock(git.IsHeadAtTag).To(func(tag string) (bool, error) {
					return false, errors.New("test error")
				}).Build()
			},
		},
		{
			name: "is-dirty-error",
			specialMock: func() {
				mockey.Mock(git.IsDirty).To(func() (bool, error) {
					return false, errors.New("test error")
				}).Build()
			},
		},
	}
	for _, tt := range tests {
		mockey.PatchConvey(tt.name, t, func() {
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
	mockey.PatchConvey("test kusion version not head tag", t, func() {
		mockGit()
		mockDependency()
		mockTime()
		mockRuntime()

		mockey.Mock(git.IsHeadAtTag).To(func(tag string) (bool, error) {
			return false, nil
		}).Build()

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
	})
}

func mockGit() {
	mockey.Mock(git.GetHeadHash).To(func() (string, error) {
		return "af79cd231e7ed1dbb00e860da9615febf5f17bf0", nil
	}).Build()
	mockey.Mock(git.GetHeadHashShort).To(func() (string, error) {
		return "af79cd23", nil
	}).Build()
	mockey.Mock(git.GetLatestTag).To(func() (string, error) {
		return "v0.3.11-alpha", nil
	}).Build()
	mockey.Mock(goversion.NewVersion).To(func(v string) (*goversion.Version, error) {
		version := &goversion.Version{}
		val := reflect.ValueOf(version).Elem().FieldByName("original")
		reflect.NewAt(val.Type(), unsafe.Pointer(val.UnsafeAddr())).Elem().Set(reflect.ValueOf("v0.3.11-alpha"))
		return version, nil
	}).Build()
	mockey.Mock(git.IsHeadAtTag).To(func(tag string) (bool, error) {
		return true, nil
	}).Build()
	mockey.Mock(git.IsDirty).To(func() (bool, error) {
		return false, nil
	}).Build()
}

func mockDependency() {
	mockey.Mock(debug.ReadBuildInfo).To(func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Deps: []*debug.Module{{Path: KclGoModulePath, Version: "stable"}, {Path: KclPluginModulePath, Version: "stable"}}}, true
	}).Build()
}

func mockTime() {
	mockey.Mock(time.Now).To(func() time.Time {
		t, _ := time.Parse("2006-01-02 15:04:05", "2006-01-02 15:04:05")
		return t
	}).Build()
}

func mockRuntime() {
	mockey.Mock(runtime.Version).To(func() string {
		return "go1.16.5"
	}).Build()
	//mockey.Mock(runtime.NumCPU).To(func() int {
	//	return 8
	//}).Build()
}

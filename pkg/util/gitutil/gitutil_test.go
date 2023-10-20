//go:build !arm64
// +build !arm64

package gitutil

import (
	"errors"
	"fmt"
	"os/exec"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
)

func TestGetRemoteURL(t *testing.T) {
	mockey.PatchConvey("get remote.origin.url", t, func() {
		url, err := getRemoteURL()
		assert.Nil(t, err)
		fmt.Println(url)
	})
	mockey.PatchConvey("cmd error", t, func() {
		mockCombinedOutput(nil, ErrMockCombinedOutput)
		_, err := getRemoteURL()
		assert.NotNil(t, err)
	})
}

func TestGetLatestTag(t *testing.T) {
	mockey.PatchConvey("get latest tag from local", t, func() {
		mockGetLatestTagFromLocal(localTag, nil)
		_, err := GetLatestTag()
		assert.Nil(t, err)
	})
	mockey.PatchConvey("get latest tag from remote", t, func() {
		mockGetLatestTagFromLocal("", ErrEmptyGitTag)
		mockGetLatestTagFromRemote(remoteTag, nil)
		_, err := GetLatestTag()
		assert.Nil(t, err)
	})
}

func TestGetLatestTagFromRemote(t *testing.T) {
	mockey.PatchConvey("remote latest tag", t, func() {
		mockGetRemoteURL(remoteURL, nil)
		tag, err := getLatestTagFromRemote()
		assert.Nil(t, err)
		fmt.Println("remote tag: ", tag)
	})
}

func TestGetLatestTagFromLocal(t *testing.T) {
	mockey.PatchConvey("get latest tag from local", t, func() {
		mockGetTagList([]string{"tag1", "tag2"}, nil)
		_, err := getLatestTagFromLocal()
		assert.Nil(t, err)
	})
	mockey.PatchConvey("error tag", t, func() {
		mockCombinedOutput(nil, ErrEmptyGitTag)
		_, err := getLatestTagFromLocal()
		assert.NotNil(t, err)
	})
}

func TestGetTagList(t *testing.T) {
	mockey.PatchConvey("cmd error", t, func() {
		mockCombinedOutput(nil, ErrMockCombinedOutput)
		_, err := getTagListFromLocal()
		assert.NotNil(t, err)
	})
	mockey.PatchConvey("cmd error", t, func() {
		_, err := getTagListFromLocal()
		assert.Nil(t, err)
	})
}

func TestGetHeadHash(t *testing.T) {
	mockey.PatchConvey("get head hash", t, func() {
		_, err := GetHeadHash()
		assert.Nil(t, err)
	})
	mockey.PatchConvey("cmd error", t, func() {
		mockCombinedOutput(nil, ErrMockCombinedOutput)
		_, err := GetHeadHash()
		assert.NotNil(t, err)
	})
}

func TestGetHeadHashShort(t *testing.T) {
	mockey.PatchConvey("get head hash error", t, func() {
		mockGetHeadHash("", ErrMockGetHeadHash)
		_, err := GetHeadHashShort()
		assert.NotNil(t, err)
	})

	mockey.PatchConvey("get head hash short", t, func() {
		mockGetHeadHash(commitSHA, nil)
		_, err := GetHeadHashShort()
		assert.Nil(t, err)
	})
}

func TestGetTagCommitSha(t *testing.T) {
	mockey.PatchConvey("error tag", t, func() {
		_, err := getTagCommitSha("")
		assert.NotNil(t, err)
	})
	mockey.PatchConvey("local tag commit sha", t, func() {
		mockGetTagCommitShaFromLocal(commitSHA, nil)
		_, err := getTagCommitSha("tag")
		assert.Nil(t, err)
	})
	mockey.PatchConvey("local tag commit sha", t, func() {
		mockGetTagCommitShaFromLocal("", ErrMockGetTagCommitShaFromLocal)
		mockGetTagCommitShaFromRemote("remote sha", nil)
		_, err := getTagCommitSha("tag")
		assert.Nil(t, err)
	})
}

func TestGetTagCommitShaFromLocal(t *testing.T) {
	mockey.PatchConvey("cmd error", t, func() {
		mockCombinedOutput(nil, ErrMockCombinedOutput)
		_, err := getTagCommitShaFromLocal("")
		assert.NotNil(t, err)
	})
}

func TestGetTagCommitShaFromRemote(t *testing.T) {
	mockey.PatchConvey("get remote.origin.url error", t, func() {
		mockGetRemoteURL("", ErrMockGetRemoteURL)
		_, err := getTagCommitShaFromRemote("")
		assert.NotNil(t, err)
	})
	mockey.PatchConvey("cmd error", t, func() {
		mockGetRemoteURL(remoteURL, nil)
		mockCombinedOutput(nil, ErrMockCombinedOutput)
		_, err := getTagCommitShaFromRemote("")
		assert.NotNil(t, err)
	})
	mockey.PatchConvey("cmd error", t, func() {
		mockGetRemoteURL(remoteURL, nil)
		_, err := getTagCommitShaFromRemote("")
		assert.Nil(t, err)
	})
}

func TestIsHeadAtTag(t *testing.T) {
	mockey.PatchConvey("empty tag", t, func() {
		_, err := IsHeadAtTag("")
		assert.NotNil(t, err)
	})
	mockey.PatchConvey("getTagCommitSha error", t, func() {
		mockGetTagCommitSha("", ErrMockGetTagCommitSha)
		_, err := IsHeadAtTag("tag")
		assert.NotNil(t, err)
	})
	mockey.PatchConvey("GetHeadHash error", t, func() {
		mockGetTagCommitSha("", nil)
		mockGetHeadHash("", ErrMockGetHeadHash)
		_, err := IsHeadAtTag("tag")
		assert.NotNil(t, err)
	})
	mockey.PatchConvey("GetHeadHash error", t, func() {
		mockGetTagCommitSha(commitSHA, nil)
		mockGetHeadHash(commitSHA, nil)
		flag, err := IsHeadAtTag("tag")
		assert.True(t, flag)
		assert.Nil(t, err)
	})
}

func TestIsDirty(t *testing.T) {
	mockey.PatchConvey("cmd err", t, func() {
		mockCombinedOutput(nil, ErrMockCombinedOutput)
		_, err := IsDirty()
		assert.NotNil(t, err)
	})
	mockey.PatchConvey("is dirty", t, func() {
		_, err := IsDirty()
		assert.Nil(t, err)
	})
}

func TestGetCurrentBranch(t *testing.T) {
	mockey.PatchConvey("cmd err", t, func() {
		mockCombinedOutput(nil, ErrMockCombinedOutput)
		_, err := GetCurrentBranch()
		assert.NotNil(t, err)
	})

	mockey.PatchConvey("success", t, func() {
		mockCombinedOutput([]byte("master"), nil)
		branch, err := GetCurrentBranch()
		assert.Nil(t, err)
		assert.Equal(t, "master", branch)
	})
}

var (
	ErrMockCombinedOutput           = errors.New("mock CombinedOutput error")
	ErrMockGetRemoteURL             = errors.New("mock getRemoteURL error")
	ErrMockGetHeadHash              = errors.New("mock GetHeadHash error")
	ErrMockGetTagCommitSha          = errors.New("mock getTagCommitSha error")
	ErrMockGetTagCommitShaFromLocal = errors.New("mock getTagCommitShaFromLocal error")
)

var (
	remoteURL = "git@github.com:KusionStack/kusion.git"
	commitSHA = "ae3518f62fa87b1bce7bc6ab2348751e558a2067"
	localTag  = "v0.3.13"
	remoteTag = "v0.3.13"
)

func mockCombinedOutput(output []byte, err error) {
	mockey.Mock((*exec.Cmd).CombinedOutput).To(func(*exec.Cmd) ([]byte, error) {
		return output, err
	}).Build()
}

func mockGetLatestTagFromLocal(tag string, err error) {
	mockey.Mock(getLatestTagFromLocal).To(func() (string, error) {
		return tag, err
	}).Build()
}

func mockGetLatestTagFromRemote(tag string, err error) {
	mockey.Mock(getLatestTagFromRemote).To(func() (string, error) {
		return tag, err
	}).Build()
}

func mockGetRemoteURL(url string, err error) {
	mockey.Mock(getRemoteURL).To(func() (string, error) {
		return url, err
	}).Build()
}

func mockGetTagList(tags []string, err error) {
	mockey.Mock(getTagListFromLocal).To(func() ([]string, error) {
		return tags, err
	}).Build()
}

func mockGetHeadHash(sha string, err error) {
	mockey.Mock(GetHeadHash).To(func() (string, error) {
		return sha, err
	}).Build()
}

func mockGetTagCommitShaFromLocal(sha string, err error) {
	mockey.Mock(getTagCommitShaFromLocal).To(func(tag string) (string, error) {
		return sha, err
	}).Build()
}

func mockGetTagCommitShaFromRemote(sha string, err error) {
	mockey.Mock(getTagCommitShaFromRemote).To(func(tag string) (string, error) {
		return sha, err
	}).Build()
}

func mockGetTagCommitSha(sha string, err error) {
	mockey.Mock(getTagCommitSha).To(func(tag string) (string, error) {
		return sha, err
	}).Build()
}

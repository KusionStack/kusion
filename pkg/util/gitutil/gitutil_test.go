package gitutil

import (
	"errors"
	"fmt"
	"os/exec"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func TestGetRemoteURL(t *testing.T) {
	t.Run("get remote.origin.url", func(t *testing.T) {
		url, err := GetRemoteURL()
		assert.Nil(t, err)
		fmt.Println(url)
	})
	t.Run("cmd error", func(t *testing.T) {
		mockCombinedOutput(nil, ErrMockCombinedOutput)
		defer monkey.UnpatchAll()
		_, err := GetRemoteURL()
		assert.NotNil(t, err)
	})
}

func TestGetLatestTag(t *testing.T) {
	t.Run("get latest tag from local", func(t *testing.T) {
		mockGetLatestTagFromLocal(localTag, nil)
		defer monkey.UnpatchAll()
		_, err := GetLatestTag()
		assert.Nil(t, err)
	})
	t.Run("get latest tag from remote", func(t *testing.T) {
		mockGetLatestTagFromLocal("", ErrEmptyGitTag)
		mockGetLatestTagFromRemote(remoteTag, nil)
		defer monkey.UnpatchAll()
		_, err := GetLatestTag()
		assert.Nil(t, err)
	})
}

func TestGetLatestTagFromRemote(t *testing.T) {
	t.Run("remote url error", func(t *testing.T) {
		mockGetRemoteURL("", ErrMockGetRemoteURL)
		defer monkey.UnpatchAll()
		_, err := GetLatestTagFromRemote()
		assert.NotNil(t, err)
	})
	t.Run("cmd error", func(t *testing.T) {
		mockGetRemoteURL("", nil)
		mockCombinedOutput(nil, ErrMockCombinedOutput)
		defer monkey.UnpatchAll()
		_, err := GetLatestTagFromRemote()
		assert.NotNil(t, err)
	})
	t.Run("remote latest tag", func(t *testing.T) {
		mockGetRemoteURL(remoteURL, nil)
		defer monkey.UnpatchAll()
		tag, err := GetLatestTagFromRemote()
		assert.Nil(t, err)
		fmt.Println("remote tag: ", tag)
	})
}

func TestGetLatestTagFromLocal(t *testing.T) {
	t.Run("get latest tag from local", func(t *testing.T) {
		mockGetTagList([]string{"tag1", "tag2"}, nil)
		defer monkey.UnpatchAll()
		_, err := GetLatestTagFromLocal()
		assert.Nil(t, err)
	})
	t.Run("error tag", func(t *testing.T) {
		mockCombinedOutput(nil, ErrEmptyGitTag)
		defer monkey.UnpatchAll()
		_, err := GetLatestTagFromLocal()
		assert.NotNil(t, err)
	})
}

func TestGetTagList(t *testing.T) {
	t.Run("cmd error", func(t *testing.T) {
		mockCombinedOutput(nil, ErrMockCombinedOutput)
		defer monkey.UnpatchAll()
		_, err := GetTagList()
		assert.NotNil(t, err)
	})
}

func TestGetHeadHash(t *testing.T) {
	t.Run("get head hash", func(t *testing.T) {
		_, err := GetHeadHash()
		assert.Nil(t, err)
	})
	t.Run("cmd error", func(t *testing.T) {
		mockCombinedOutput(nil, ErrMockCombinedOutput)
		defer monkey.UnpatchAll()
		_, err := GetHeadHash()
		assert.NotNil(t, err)
	})
}

func TestGetHeadHashShort(t *testing.T) {
	t.Run("get head hash error", func(t *testing.T) {
		mockGetHeadHash("", ErrMockGetHeadHash)
		defer monkey.UnpatchAll()
		_, err := GetHeadHashShort()
		assert.NotNil(t, err)
	})

	t.Run("get head hash short", func(t *testing.T) {
		mockGetHeadHash(commitSHA, nil)
		defer monkey.UnpatchAll()
		_, err := GetHeadHashShort()
		assert.Nil(t, err)
	})
}

func TestGetTagCommitSha(t *testing.T) {
	t.Run("error tag", func(t *testing.T) {
		_, err := GetTagCommitSha("")
		assert.NotNil(t, err)
	})
	t.Run("local tag commit sha", func(t *testing.T) {
		mockGetTagCommitShaFromLocal(commitSHA, nil)
		defer monkey.UnpatchAll()
		_, err := GetTagCommitSha("tag")
		assert.Nil(t, err)
	})
	t.Run("local tag commit sha", func(t *testing.T) {
		mockGetTagCommitShaFromLocal("", ErrMockGetTagCommitShaFromLocal)
		mockGetTagCommitShaFromRemote("remote sha", nil)
		defer monkey.UnpatchAll()
		_, err := GetTagCommitSha("tag")
		assert.Nil(t, err)
	})
}

func TestGetTagCommitShaFromLocal(t *testing.T) {
	t.Run("cmd error", func(t *testing.T) {
		mockCombinedOutput(nil, ErrMockCombinedOutput)
		defer monkey.UnpatchAll()
		_, err := GetTagCommitShaFromLocal("")
		assert.NotNil(t, err)
	})
}

func TestGetTagCommitShaFromRemote(t *testing.T) {
	t.Run("get remote.origin.url error", func(t *testing.T) {
		mockGetRemoteURL("", ErrMockGetRemoteURL)
		defer monkey.UnpatchAll()
		_, err := GetTagCommitShaFromRemote("")
		assert.NotNil(t, err)
	})
	t.Run("cmd error", func(t *testing.T) {
		mockGetRemoteURL(remoteURL, nil)
		mockCombinedOutput(nil, ErrMockCombinedOutput)
		defer monkey.UnpatchAll()
		_, err := GetTagCommitShaFromRemote("")
		assert.NotNil(t, err)
	})
	t.Run("cmd error", func(t *testing.T) {
		mockGetRemoteURL(remoteURL, nil)
		defer monkey.UnpatchAll()
		_, err := GetTagCommitShaFromRemote("")
		assert.Nil(t, err)
	})
}

func TestIsHeadAtTag(t *testing.T) {
	t.Run("empty tag", func(t *testing.T) {
		_, err := IsHeadAtTag("")
		assert.NotNil(t, err)
	})
	t.Run("GetTagCommitSha error", func(t *testing.T) {
		mockGetTagCommitSha("", ErrMockGetTagCommitSha)
		defer monkey.UnpatchAll()
		_, err := IsHeadAtTag("tag")
		assert.NotNil(t, err)
	})
	t.Run("GetHeadHash error", func(t *testing.T) {
		mockGetTagCommitSha("", nil)
		mockGetHeadHash("", ErrMockGetHeadHash)
		defer monkey.UnpatchAll()
		_, err := IsHeadAtTag("tag")
		assert.NotNil(t, err)
	})
	t.Run("GetHeadHash error", func(t *testing.T) {
		mockGetTagCommitSha(commitSHA, nil)
		mockGetHeadHash(commitSHA, nil)
		defer monkey.UnpatchAll()
		flag, err := IsHeadAtTag("tag")
		assert.True(t, flag)
		assert.Nil(t, err)
	})
}

func TestIsDirty(t *testing.T) {
	t.Run("cmd err", func(t *testing.T) {
		mockCombinedOutput(nil, ErrMockCombinedOutput)
		defer monkey.UnpatchAll()
		_, err := IsDirty()
		assert.NotNil(t, err)
	})
	t.Run("is dirty", func(t *testing.T) {
		_, err := IsDirty()
		assert.Nil(t, err)
	})
}

func TestGetCurrentBranch(t *testing.T) {
	t.Run("cmd err", func(t *testing.T) {
		mockCombinedOutput(nil, ErrMockCombinedOutput)
		defer monkey.UnpatchAll()
		_, err := GetCurrentBranch()
		assert.NotNil(t, err)
	})

	t.Run("success", func(t *testing.T) {
		mockCombinedOutput([]byte("master"), nil)
		defer monkey.UnpatchAll()
		branch, err := GetCurrentBranch()
		assert.Nil(t, err)
		assert.Equal(t, "master", branch)
	})
}

var (
	ErrMockCombinedOutput           = errors.New("mock CombinedOutput error")
	ErrMockGetRemoteURL             = errors.New("mock GetRemoteURL error")
	ErrMockGetHeadHash              = errors.New("mock GetHeadHash error")
	ErrMockGetTagCommitSha          = errors.New("mock GetTagCommitSha error")
	ErrMockGetTagCommitShaFromLocal = errors.New("mock GetTagCommitShaFromLocal error")
)

var (
	remoteURL = "git@github.com:KusionStack/kusion.git"
	commitSHA = "ae3518f62fa87b1bce7bc6ab2348751e558a2067"
	localTag  = "v0.3.13"
	remoteTag = "v0.3.13"
)

func mockCombinedOutput(output []byte, err error) {
	monkey.Patch((*exec.Cmd).CombinedOutput, func(*exec.Cmd) ([]byte, error) {
		return output, err
	})
}

func mockGetLatestTagFromLocal(tag string, err error) {
	monkey.Patch(GetLatestTagFromLocal, func() (string, error) {
		return tag, err
	})
}

func mockGetLatestTagFromRemote(tag string, err error) {
	monkey.Patch(GetLatestTagFromRemote, func() (string, error) {
		return tag, err
	})
}

func mockGetRemoteURL(url string, err error) {
	monkey.Patch(GetRemoteURL, func() (string, error) {
		return url, err
	})
}

func mockGetTagList(tags []string, err error) {
	monkey.Patch(GetTagList, func() ([]string, error) {
		return tags, err
	})
}

func mockGetHeadHash(sha string, err error) {
	monkey.Patch(GetHeadHash, func() (string, error) {
		return sha, err
	})
}

func mockGetTagCommitShaFromLocal(sha string, err error) {
	monkey.Patch(GetTagCommitShaFromLocal, func(tag string) (string, error) {
		return sha, err
	})
}

func mockGetTagCommitShaFromRemote(sha string, err error) {
	monkey.Patch(GetTagCommitShaFromRemote, func(tag string) (string, error) {
		return sha, err
	})
}

func mockGetTagCommitSha(sha string, err error) {
	monkey.Patch(GetTagCommitSha, func(tag string) (string, error) {
		return sha, err
	})
}

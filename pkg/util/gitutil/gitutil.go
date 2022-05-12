package gitutil

import (
	"errors"
	"os/exec"
	"strings"
)

// https://git-scm.com/docs/git-tag
// https://github.com/vivin/better-setuptools-git-version/blob/master/better_setuptools_git_version.py

var ErrEmptyGitTag = errors.New("empty tag")

// get remote url
func GetRemoteURL() (string, error) {
	stdout, err := exec.Command(
		"git", "config", "--get", "remote.origin.url",
	).CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(stdout)), nil
}

func GetLatestTag() (string, error) {
	tag, err := GetLatestTagFromLocal()
	if tag == "" || err != nil {
		return GetLatestTagFromRemote()
	}
	return tag, nil
}

// get latest tag from remote,
// the fitting git clone depth is 1
func GetLatestTagFromRemote() (tag string, err error) {
	// get remote url
	remoteURL, err := GetRemoteURL()
	if err != nil {
		return "", err
	}

	// get latest tag from remote
	stdout, err := exec.Command(
		`bash`, `-c`, `git ls-remote --tags --sort=v:refname `+remoteURL+` | tail -n1 | sed 's/.*\///; s/\^{}//'`,
	).CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(stdout)), nil
}

func GetLatestTagFromLocal() (tag string, err error) {
	tags, err := GetTagList()
	if err != nil {
		return "", err
	}
	if len(tags) > 0 {
		tag = tags[len(tags)-1]
	}
	return strings.TrimSpace(tag), nil
}

func GetTagList() (tags []string, err error) {
	// git tag --merged
	stdout, err := exec.Command(
		`git`, `describe`, `--abbrev=0`, `--tags`,
	).CombinedOutput()
	if err != nil {
		return nil, err
	}

	for _, s := range strings.Split(strings.TrimSpace(string(stdout)), "\n") {
		if s := strings.TrimSpace(s); s != "" {
			tags = append(tags, s)
		}
	}
	return
}

func GetTagListFromRemote(remoteURL string, reverse bool) (tags []string, err error) {
	tmpTags := []string{}
	// Get all tags from remote
	stdout, err := exec.Command(
		`bash`, `-c`, `git ls-remote --tags --sort=v:refname `+remoteURL+` | sed 's/.*\///; s/\^{}//'`,
	).CombinedOutput()
	if err != nil {
		return nil, err
	}

	for _, s := range strings.Split(strings.TrimSpace(string(stdout)), "\n") {
		if s := strings.TrimSpace(s); s != "" {
			tmpTags = append(tmpTags, s)
		}
	}

	// Reverse slice
	if reverse {
		for i, j := 0, len(tmpTags)-1; i < j; i, j = i+1, j-1 {
			tmpTags[i], tmpTags[j] = tmpTags[j], tmpTags[i]
		}
	}

	// Remove duplicates
	tagSet := make(map[string]struct{})
	for _, tag := range tmpTags {
		if _, ok := tagSet[tag]; !ok {
			tags = append(tags, tag)
			tagSet[tag] = struct{}{}
		}
	}
	return
}

func GetHeadHash() (sha string, err error) {
	// git rev-parse HEAD
	stdout, err := exec.Command(
		`git`, `rev-parse`, `HEAD`,
	).CombinedOutput()
	if err != nil {
		return "", err
	}

	sha = strings.TrimSpace(string(stdout))
	return
}

func GetHeadHashShort() (sha string, err error) {
	sha, err = GetHeadHash()
	if err != nil {
		return "", err
	}
	if len(sha) > 8 {
		sha = sha[:8]
	}
	return
}

func GetTagCommitSha(tag string) (sha string, err error) {
	if tag == "" {
		return "", ErrEmptyGitTag
	}

	sha, err = GetTagCommitShaFromLocal(tag)
	if sha == "" || err != nil {
		return GetTagCommitShaFromRemote(tag)
	}
	return
}

func GetTagCommitShaFromLocal(tag string) (sha string, err error) {
	// git rev-list -n 1 {tag}
	stdout, err := exec.Command(
		`git`, `rev-list`, `-n`, `1`, tag,
	).CombinedOutput()
	if err != nil {
		return "", err
	}
	var lines []string
	for _, s := range strings.Split(strings.TrimSpace(string(stdout)), "\n") {
		if s := strings.TrimSpace(s); s != "" {
			lines = append(lines, s)
		}
	}
	if len(lines) > 0 {
		sha = lines[len(lines)-1]
	}
	return strings.TrimSpace(sha), nil
}

// get tag commit sha from remote,
// the fitting git clone depth is 1
func GetTagCommitShaFromRemote(_ string) (string, error) {
	// get remote url
	remoteURL, err := GetRemoteURL()
	if err != nil {
		return "", err
	}

	stdout, err := exec.Command(
		`bash`, `-c`, `git ls-remote --tags --sort=v:refname `+remoteURL+` | tail -n1 | awk '{print $1}'`,
	).CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(stdout)), nil
}

func IsHeadAtTag(tag string) (bool, error) {
	if tag == "" {
		return false, ErrEmptyGitTag
	}
	sha1, err1 := GetTagCommitSha(tag)
	if err1 != nil {
		return false, err1
	}
	sha2, err2 := GetHeadHash()
	if err2 != nil {
		return false, err2
	}
	return sha1 == sha2, nil
}

func IsDirty() (dirty bool, err error) {
	// git status -s
	stdout, err := exec.Command(
		`git`, `status`, `-s`,
	).CombinedOutput()
	if err != nil {
		return false, err
	}

	dirty = strings.TrimSpace(string(stdout)) != ""
	return
}

func GetCurrentBranch() (string, error) {
	// git status -s
	stdout, err := exec.Command(
		`git`, `symbolic-ref`, `--short`, `-q`, `HEAD`,
	).CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(stdout)), nil
}

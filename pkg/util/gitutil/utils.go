package gitutil

import (
	"context"
	"errors"
	"os/exec"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/google/go-github/v50/github"
)

// https://git-scm.com/docs/git-tag
// https://github.com/vivin/better-setuptools-git-version/blob/master/better_setuptools_git_version.py

var ErrEmptyGitTag = errors.New("empty tag")

const (
	Owner = "KusionStack"
	Repo  = "kusion"
)

func GetLatestTag() (string, error) {
	tag, err := getLatestTagFromLocal()
	if tag == "" || err != nil {
		return getLatestTagFromRemote()
	}
	return tag, nil
}

// getLatestTagFromRemote the fitting git clone depth is 1
func getLatestTagFromRemote() (string, error) {
	tags, err := getTagListFromRemote(Owner, Repo)
	if err != nil {
		return "", err
	}

	return tags[0], nil
}

// get remote url
func getRemoteURL() (string, error) {
	stdout, err := exec.Command(
		"git", "config", "--get", "remote.origin.url",
	).CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(stdout)), nil
}

func getLatestTagFromLocal() (tag string, err error) {
	tags, err := getTagListFromLocal()
	if err != nil {
		return "", err
	}
	if len(tags) > 0 {
		tag = tags[len(tags)-1]
	}
	return strings.TrimSpace(tag), nil
}

func getTagListFromLocal() (tags []string, err error) {
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

func getTagListFromRemote(owner, repo string) (tags []string, err error) {
	client := github.NewClient(nil)
	rts, _, err := client.Repositories.ListTags(context.Background(), owner, repo, nil)
	if err != nil {
		return nil, err
	}
	if len(rts) == 0 {
		return nil, errors.New("no tag found")
	}

	for _, rt := range rts {
		v, err := semver.ParseTolerant(rt.GetName())
		if err != nil {
			continue
		}
		if len(v.Pre) > 0 || len(v.Build) > 0 {
			continue
		}
		tags = append(tags, rt.GetName())
	}

	return tags, nil
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

func IsHeadAtTag(tag string) (bool, error) {
	if tag == "" {
		return false, ErrEmptyGitTag
	}
	sha1, err1 := getTagCommitSha(tag)
	if err1 != nil {
		return false, err1
	}
	sha2, err2 := GetHeadHash()
	if err2 != nil {
		return false, err2
	}
	return sha1 == sha2, nil
}

func getTagCommitSha(tag string) (sha string, err error) {
	if tag == "" {
		return "", ErrEmptyGitTag
	}

	sha, err = getTagCommitShaFromLocal(tag)
	if sha == "" || err != nil {
		return getTagCommitShaFromRemote(tag)
	}
	return
}

func getTagCommitShaFromLocal(tag string) (sha string, err error) {
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
func getTagCommitShaFromRemote(_ string) (string, error) {
	// get remote url
	remoteURL, err := getRemoteURL()
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

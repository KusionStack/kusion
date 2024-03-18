package gitutil

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Info contains basic information of Git repository.
type Info struct {
	RemoteURL  string
	Commit     string
	CommitDate string
}

// Get returns the overall codebase version.
func Get(repoRoot string) (info Info) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tsCmd := exec.CommandContext(ctx, "git", "--no-pager", "log", "-1", `--format=%ct`)
	tsCmd.Dir = repoRoot
	if ts, err := tsCmd.Output(); err == nil && len(ts) > 1 {
		if i, err := strconv.ParseInt(strings.TrimSuffix(string(ts), "\n"), 10, 64); err == nil {
			d := time.Unix(i, 0)
			info.CommitDate = d.Format(time.RFC3339)
		}
	}

	urlCmd := exec.CommandContext(ctx, "git", "config", "--get", "remote.origin.url")
	urlCmd.Dir = repoRoot
	if repo, err := urlCmd.Output(); err == nil && len(repo) > 1 {
		info.RemoteURL = strings.TrimSuffix(string(repo), "\n")
	}

	shaCmd := exec.CommandContext(ctx, "git", "show", "-s", "--format=%H")
	shaCmd.Dir = repoRoot
	if commit, err := shaCmd.Output(); err == nil && len(commit) > 1 {
		info.Commit = strings.TrimSuffix(string(commit), "\n")
	}

	return info
}

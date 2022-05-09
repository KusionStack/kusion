package gitutil

import "os"

// GitInfo contains git information.
type GitInfo struct {
	CurrentBranch string `json:"currentBranch,omitempty" yaml:"currentBranch,omitempty"` // Such as "master"
	HeadCommit    string `json:"headCommit,omitempty" yaml:"headCommit,omitempty"`       // Such as "3836f8770ab8f488356b2129f42f2ae5c1134bb0"
	TreeState     string `json:"treeState,omitempty" yaml:"treeState,omitempty"`         // Such as "clean", "dirty"
	LatestTag     string `json:"latestTag,omitempty" yaml:"latestTag,omitempty"`         // Such as "v1.2.3"
}

// NewGitInfoFrom returns git info from workDir, or nil if failed.
func NewGitInfoFrom(workDir string) (*GitInfo, error) {
	// Cd to workDir
	oldWd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	_ = os.Chdir(workDir)
	defer os.Chdir(oldWd)

	// Get git info
	var (
		curCommit    string
		curBranch    string
		latestTag    string
		isDirty      bool
		gitTreeState string
	)

	if curCommit, err = GetHeadHash(); err != nil {
		return nil, err
	}

	if curBranch, err = GetCurrentBranch(); err != nil {
		return nil, err
	}

	if latestTag, err = GetLatestTag(); err != nil {
		return nil, err
	}

	if isDirty, err = IsDirty(); err != nil {
		return nil, err
	}

	// Get git tree state
	if isDirty {
		gitTreeState = "dirty"
	} else {
		gitTreeState = "clean"
	}

	return &GitInfo{
		LatestTag:     latestTag,
		CurrentBranch: curBranch,
		HeadCommit:    curCommit,
		TreeState:     gitTreeState,
	}, nil
}

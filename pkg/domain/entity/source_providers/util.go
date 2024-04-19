package sourceproviders

import (
	"fmt"

	"github.com/go-git/go-git/v5" // with go modules enabled (GO111MODULE=on or outside GOPATH)
	"github.com/go-git/go-git/v5/plumbing"
)

// func switchBranch(repoPath string, branchName string) error {
// 	// Open an existing repository
// 	r, err := git.PlainOpen(repoPath)
// 	if err != nil {
// 		return err
// 	}

// 	// Get the worktree for the repository
// 	w, err := r.Worktree()
// 	if err != nil {
// 		return err
// 	}

// 	// Checkout the specified branch
// 	err = w.Checkout(&git.CheckoutOptions{
// 		Branch: plumbing.NewBranchReferenceName(branchName),
// 	})

// 	if err != nil {
// 		return err
// 	}

// 	fmt.Println("Switched to branch:", branchName)
// 	return nil
// }

func checkoutRevision(repoPath string, revision string) error {
	// Open an existing repository
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	// Get the worktree for the repository
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	// Checkout the specified revision
	// For a commit or tag, use `plumbing.Revision` to resolve the hash
	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(revision),
	})
	if err != nil {
		return err
	}

	fmt.Println("Checked out to revision:", revision)
	return nil
}

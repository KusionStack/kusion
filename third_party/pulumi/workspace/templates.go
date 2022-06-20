// Copyright 2016-2018, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"kusionstack.io/kusion/third_party/pulumi/gitutil"
)

// RetrieveGitFolder downloads the repo to path and returns the full path on disk.
func RetrieveGitFolder(rawurl string, path string) (string, error) {
	url, urlPath, err := gitutil.ParseGitRepoURL(rawurl)
	if err != nil {
		return "", err
	}

	ref, commit, subDirectory, err := gitutil.GetGitReferenceNameOrHashAndSubDirectory(url, urlPath)
	if err != nil {
		return "", fmt.Errorf("failed to get git ref: %w", err)
	}
	if ref != "" {

		// Different reference attempts to cycle through
		// We default to master then main in that order. We need to order them to avoid breaking
		// already existing processes for repos that already have a master and main branch.
		refAttempts := []plumbing.ReferenceName{plumbing.Master, plumbing.NewBranchReferenceName("main")}

		if ref != plumbing.HEAD {
			// If we have a non-default reference, we just use it
			refAttempts = []plumbing.ReferenceName{ref}
		}

		var cloneErr error
		for _, ref := range refAttempts {
			// Attempt the clone. If it succeeds, break
			cloneErr := gitutil.GitCloneOrPull(url, ref, path, true /*shallow*/)
			if cloneErr == nil {
				break
			}
		}
		if cloneErr != nil {
			return "", fmt.Errorf("failed to clone ref '%s': %w", refAttempts[len(refAttempts)-1], cloneErr)
		}

	} else {
		if cloneErr := gitutil.GitCloneAndCheckoutCommit(url, commit, path); cloneErr != nil {
			return "", fmt.Errorf("failed to clone and checkout %s(%s): %w", url, commit, cloneErr)
		}
	}

	// Verify the sub directory exists.
	fullPath := filepath.Join(path, filepath.FromSlash(subDirectory))
	info, err := os.Stat(fullPath)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.Errorf("%s is not a directory", fullPath)
	}

	return fullPath, nil
}

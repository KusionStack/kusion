// Copyright 2024 KusionStack Authors
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

package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Create returns a suitable location relative to which the file with the
// specified `name` can be written. The first path from the provided `paths`
// slice which is successfully created (or already exists) is used as a base
// path for the file. The `name` parameter should contain the name of the file
// which is going to be written in the location returned by this function, but
// it can also contain a set of parent directories, which will be created
// relative to the selected parent path.
func Create(name string, paths []string) (string, error) {
	searchedPaths := make([]string, len(paths))
	for _, p := range paths {
		p = filepath.Join(p, name)

		dir := filepath.Dir(p)
		if Exists(dir) {
			return p, nil
		}
		if err := os.MkdirAll(dir, os.ModeDir|0o700); err == nil {
			return p, nil
		}

		searchedPaths = append(searchedPaths, dir)
	}

	return "", fmt.Errorf("could not create any of the following paths: %s",
		strings.Join(searchedPaths, ", "))
}

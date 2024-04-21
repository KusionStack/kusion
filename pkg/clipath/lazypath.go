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

package clipath

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"

	"kusionstack.io/kusion/pkg/util/pathutil"
)

const (
	// CacheHomeEnvVar is the environment variable used by Kusion
	// for the cache directory. When no value is set a default is used.
	CacheHomeEnvVar = "KUSION_CACHE_HOME"

	// ConfigHomeEnvVar is the environment variable used by Kusion
	// for the config directory. When no value is set a default is used.
	ConfigHomeEnvVar = "KUSION_CONFIG_HOME"

	// DataHomeEnvVar is the environment variable used by Kusion
	// for the data directory. When no value is set a default is used.
	DataHomeEnvVar = "KUSION_DATA_HOME"

	// StateHomeEnvVar is the environment variable used by Kusion
	// for the state data directory. When no value is set a default is used.
	StateHomeEnvVar = "KUSION_STATE_HOME"
)

// lazyPath is a lazy-loaded path buffer for the XDG base directory specification.
type lazyPath string

func (l lazyPath) createPath(kusionEnvVar string, xdgPaths []string, elem ...string) (string, error) {
	// Construct paths with following order:
	// 1. See if a Kusion specific environment variable has been set.
	// 2. Append provided XDG paths
	var paths []string

	base := os.Getenv(kusionEnvVar)
	if base != "" {
		paths = append(paths, base)
	}

	paths = append(paths, xdgPaths...)

	return pathutil.Create(filepath.Join(string(l), filepath.Join(elem...)), paths)
}

// configPath returns a suitable location where user specific configuration files should be stored.
func (l lazyPath) configPath(elem ...string) (string, error) {
	return l.createPath(ConfigHomeEnvVar, append([]string{xdg.ConfigHome}, xdg.ConfigDirs...), elem...)
}

// cachePath returns a suitable location where user specific non-essential data files should be stored.
func (l lazyPath) cachePath(elem ...string) (string, error) {
	return l.createPath(CacheHomeEnvVar, []string{xdg.CacheHome}, elem...)
}

// dataPath returns a suitable location where user specific data files should be stored.
func (l lazyPath) dataPath(elem ...string) (string, error) {
	return l.createPath(DataHomeEnvVar, append([]string{xdg.DataHome}, xdg.DataDirs...), elem...)
}

// statePath returns a suitable location where user specific state files should be stored.
func (l lazyPath) statePath(elem ...string) (string, error) {
	return l.createPath(StateHomeEnvVar, []string{xdg.StateHome}, elem...)
}

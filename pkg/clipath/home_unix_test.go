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

//go:build !windows

package clipath

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
)

func TestKusionHome(t *testing.T) {
	tmpCacheDir := filepath.Join(os.TempDir(), "cache")
	tmpConfigDir := filepath.Join(os.TempDir(), "config")
	tmpDataDir := filepath.Join(os.TempDir(), "data")
	tmpStateDir := filepath.Join(os.TempDir(), "state")

	os.Setenv("XDG_CACHE_HOME", tmpCacheDir)
	os.Setenv("XDG_CONFIG_HOME", tmpConfigDir)
	os.Setenv("XDG_DATA_HOME", tmpDataDir)
	os.Setenv("XDG_STATE_HOME", tmpStateDir)

	xdg.Reload()

	cachePath, err := CachePath()
	assert.Nil(t, err)
	assert.Equal(t, filepath.Join(tmpCacheDir, string(lp)), cachePath)

	configPath, err := ConfigPath()
	assert.Nil(t, err)
	assert.Equal(t, filepath.Join(tmpConfigDir, string(lp)), configPath)

	dataPath, err := DataPath()
	assert.Nil(t, err)
	assert.Equal(t, filepath.Join(tmpDataDir, string(lp)), dataPath)

	statePath, err := StatePath()
	assert.Nil(t, err)
	assert.Equal(t, filepath.Join(tmpStateDir, string(lp)), statePath)
}

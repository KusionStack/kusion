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

//go:build darwin

package clipath

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
)

const (
	appName  = "kusion"
	testFile = "test.txt"
	lazy     = lazyPath(appName)
)

func TestCachePath(t *testing.T) {
	actual, err := lazy.cachePath(testFile)
	assert.Nil(t, err)

	homeDir, err := os.UserHomeDir()
	assert.Nil(t, err)

	expected := filepath.Join(homeDir, "Library", "Caches", appName, testFile)
	assert.Equal(t, expected, actual)

	os.Setenv("XDG_CACHE_HOME", "/tmp")

	xdg.Reload()

	actual, err = lazy.cachePath(testFile)
	assert.Nil(t, err)

	expected = filepath.Join("/tmp", appName, testFile)
	assert.Equal(t, expected, actual)
}

func TestConfigPath(t *testing.T) {
	actual, err := lazy.configPath(testFile)
	assert.Nil(t, err)

	homeDir, err := os.UserHomeDir()
	assert.Nil(t, err)

	expected := filepath.Join(homeDir, "Library", "Application Support", appName, testFile)
	assert.Equal(t, expected, actual)

	os.Setenv("XDG_CONFIG_HOME", "/tmp")

	xdg.Reload()

	actual, err = lazy.configPath(testFile)
	assert.Nil(t, err)

	expected = filepath.Join("/tmp", appName, testFile)
	assert.Equal(t, expected, actual)
}

func TestDataPath(t *testing.T) {
	actual, err := lazy.dataPath(testFile)
	assert.Nil(t, err)

	homeDir, err := os.UserHomeDir()
	assert.Nil(t, err)

	expected := filepath.Join(homeDir, "Library", "Application Support", appName, testFile)
	assert.Equal(t, expected, actual)

	os.Setenv("XDG_DATA_HOME", "/tmp")

	xdg.Reload()

	actual, err = lazy.dataPath(testFile)
	assert.Nil(t, err)

	expected = filepath.Join("/tmp", appName, testFile)
	assert.Equal(t, expected, actual)
}

func TestStatePath(t *testing.T) {
	actual, err := lazy.statePath(testFile)
	assert.Nil(t, err)

	homeDir, err := os.UserHomeDir()
	assert.Nil(t, err)

	expected := filepath.Join(homeDir, "Library", "Application Support", appName, testFile)
	assert.Equal(t, expected, actual)

	os.Setenv("XDG_STATE_HOME", "/tmp")

	xdg.Reload()

	actual, err = lazy.statePath(testFile)
	assert.Nil(t, err)

	expected = filepath.Join("/tmp", appName, testFile)
	assert.Equal(t, expected, actual)
}

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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExists(t *testing.T) {
	tempDir := os.TempDir()

	// Test regular file.
	pathFile := filepath.Join(tempDir, "regular")
	f, err := os.Create(pathFile)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	require.True(t, Exists(pathFile))

	// Test symlink.
	pathSymlink := filepath.Join(tempDir, "symlink")
	require.NoError(t, os.Symlink(pathFile, pathSymlink))
	require.True(t, Exists(pathSymlink))

	// Test non-existent file.
	require.NoError(t, os.Remove(pathFile))
	require.False(t, Exists(pathFile))
	require.False(t, Exists(pathSymlink))
	require.NoError(t, os.Remove(pathSymlink))
	require.False(t, Exists(pathSymlink))
}

func TestCreate(t *testing.T) {
	tempDir := os.TempDir()

	// Test path selection order.
	p, err := Create("test", []string{tempDir, "\000a"})
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tempDir, "test"), p)

	p, err = Create("test", []string{"\000a", tempDir})
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tempDir, "test"), p)

	// Test relative parent directories.
	expected := filepath.Join(tempDir, "kusion", "config", "test")
	p, err = Create(filepath.Join("kusion", "config", "test"), []string{"\000a", tempDir})
	require.NoError(t, err)
	require.Equal(t, expected, p)
	require.NoError(t, os.RemoveAll(filepath.Dir(expected)))

	expected = filepath.Join(tempDir, "kusion", "test")
	p, err = Create(filepath.Join("kusion", "test"), []string{"\000a", tempDir})
	require.NoError(t, err)
	require.Equal(t, expected, p)
	require.NoError(t, os.RemoveAll(filepath.Dir(expected)))

	// Test invalid paths.
	_, err = Create(filepath.Join("kusion", "test"), []string{"\000a"})
	require.Error(t, err)

	_, err = Create("test", []string{filepath.Join(tempDir, "\000a")})
	require.Error(t, err)
}

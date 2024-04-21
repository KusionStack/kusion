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

//go:build windows

package clipath

func TestKusionHome(t *testing.T) {
	os.Setenv("XDG_CACHE_HOME", "c:\\")
	os.Setenv("XDG_CONFIG_HOME", "d:\\")
	os.Setenv("XDG_DATA_HOME", "e:\\")
	os.Setenv("XDG_STATE_HOME", "f:\\")

	xdg.Reload()

	cachePath, err := CachePath()
	assert.Nil(t, err)
	assert.Equal(t, "c:\\kusion", cachePath)

	configPath, err := ConfigPath()
	assert.Nil(t, err)
	assert.Equal(t, "d:\\kusion", configPath)

	dataPath, err := DataPath()
	assert.Nil(t, err)
	assert.Equal(t, "e:\\kusion", dataPath)

	statePath, err := StatePath()
	assert.Nil(t, err)
	assert.Equal(t, "f:\\kusion", statePath)
}

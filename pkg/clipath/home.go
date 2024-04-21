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

// Package clipath calculates filesystem paths to Kusion's configuration, cache and data.
package clipath

// This helper builds paths to Kusion's configuration, cache and data paths.
const lp = lazyPath("kusion")

// ConfigPath returns the path where Kusion stores configuration.
func ConfigPath(elem ...string) (string, error) { return lp.configPath(elem...) }

// CachePath returns the path where Kusion stores cached objects.
func CachePath(elem ...string) (string, error) { return lp.cachePath(elem...) }

// DataPath returns the path where Kusion stores data.
func DataPath(elem ...string) (string, error) { return lp.dataPath(elem...) }

// StatePath returns the path where Kusion stores state data (logs, history, recently used files e.g.).
func StatePath(elem ...string) (string, error) { return lp.statePath(elem...) }

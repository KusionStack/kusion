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

package v1

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
)

func (r *Resource) ResourceKey() string {
	return r.ID
}

func (rs Resources) Index() map[string]*Resource {
	m := make(map[string]*Resource)
	for i := range rs {
		m[rs[i].ResourceKey()] = &rs[i]
	}
	return m
}

// GVKIndex returns a map of GVK to resources, for now, only Kubernetes resources.
func (rs Resources) GVKIndex() map[string][]*Resource {
	m := make(map[string][]*Resource)
	for i := range rs {
		resource := &rs[i]
		if resource.Type != Kubernetes {
			continue
		}
		gvk := resource.Extensions[ResourceExtensionGVK].(string)
		m[gvk] = append(m[gvk], resource)
	}
	return m
}

func (rs Resources) Len() int      { return len(rs) }
func (rs Resources) Swap(i, j int) { rs[i], rs[j] = rs[j], rs[i] }
func (rs Resources) Less(i, j int) bool {
	switch {
	case rs[i].ID != rs[j].ID:
		return rs[i].ID < rs[j].ID
	default:
		return false
	}
}

func (r *Resource) DeepCopy() (*Resource, error) {
	if r == nil {
		return nil, fmt.Errorf("source resource is nil")
	}
	data, err := jsoniter.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource: %v", err)
	}
	var res Resource
	err = jsoniter.Unmarshal(data, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal resource: %v", err)
	}
	return &res, nil
}

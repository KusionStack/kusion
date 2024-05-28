package storages

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func testDataFolder(releasePath string) string {
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, "testdata", releasePath, "test_project", "test_ws")
}

func mockRelease(revision uint64) *v1.Release {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return &v1.Release{
		Project:   "test_project",
		Workspace: "test_ws",
		Revision:  revision,
		Stack:     "test_stack",
		Spec: &v1.Spec{
			Resources: v1.Resources{
				v1.Resource{
					ID:   "apps.kusionstack.io/v1alpha1:PodTransitionRule:fakeNs:default-dev-foo",
					Type: "Kubernetes",
					Attributes: map[string]interface{}{
						"apiVersion": "apps.kusionstack.io/v1alpha1",
						"kind":       "PodTransitionRule",
						"metadata": map[string]interface{}{
							"creationTimestamp": interface{}(nil),
							"name":              "default-dev-foo",
							"namespace":         "fakeNs",
						},
						"spec": map[string]interface{}{
							"rules": []interface{}{map[string]interface{}{
								"availablePolicy": map[string]interface{}{
									"maxUnavailableValue": "30%",
								},
								"name": "maxUnavailable",
							}},
							"selector": map[string]interface{}{
								"matchLabels": map[string]interface{}{
									"app.kubernetes.io/name": "foo", "app.kubernetes.io/part-of": "default",
								},
							},
						}, "status": map[string]interface{}{},
					},
					DependsOn: []string(nil),
					Extensions: map[string]interface{}{
						"GVK": "apps.kusionstack.io/v1alpha1, Kind=PodTransitionRule",
					},
				},
			},
		},
		State: &v1.State{
			Resources: v1.Resources{
				v1.Resource{
					ID:   "apps.kusionstack.io/v1alpha1:PodTransitionRule:fakeNs:default-dev-foo",
					Type: "Kubernetes",
					Attributes: map[string]interface{}{
						"apiVersion": "apps.kusionstack.io/v1alpha1",
						"kind":       "PodTransitionRule",
						"metadata": map[string]interface{}{
							"creationTimestamp": interface{}(nil),
							"name":              "default-dev-foo",
							"namespace":         "fakeNs",
						},
						"spec": map[string]interface{}{
							"rules": []interface{}{map[string]interface{}{
								"availablePolicy": map[string]interface{}{
									"maxUnavailableValue": "30%",
								},
								"name": "maxUnavailable",
							}},
							"selector": map[string]interface{}{
								"matchLabels": map[string]interface{}{
									"app.kubernetes.io/name": "foo", "app.kubernetes.io/part-of": "default",
								},
							},
						}, "status": map[string]interface{}{},
					},
					DependsOn: []string(nil),
					Extensions: map[string]interface{}{
						"GVK": "apps.kusionstack.io/v1alpha1, Kind=PodTransitionRule",
					},
				},
			},
		},
		Phase:        v1.ReleasePhaseSucceeded,
		CreateTime:   time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
		ModifiedTime: time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
	}
}

func mockReleaseRevision1Content() string {
	return `
project: test_project
workspace: test_ws
revision: 1
stack: test_stack
spec:
  resources:
    - id: apps.kusionstack.io/v1alpha1:PodTransitionRule:fakeNs:default-dev-foo
      type: Kubernetes
      attributes:
        apiVersion: apps.kusionstack.io/v1alpha1
        kind: PodTransitionRule
        metadata:
          creationTimestamp: null
          name: default-dev-foo
          namespace: fakeNs
        spec:
          rules:
            - availablePolicy:
                maxUnavailableValue: 30%
              name: maxUnavailable
          selector:
            matchLabels:
              app.kubernetes.io/name: foo
              app.kubernetes.io/part-of: default
        status: {}
      extensions:
        GVK: apps.kusionstack.io/v1alpha1, Kind=PodTransitionRule
state:
  resources:
    - id: apps.kusionstack.io/v1alpha1:PodTransitionRule:fakeNs:default-dev-foo
      type: Kubernetes
      attributes:
        apiVersion: apps.kusionstack.io/v1alpha1
        kind: PodTransitionRule
        metadata:
          creationTimestamp: null
          name: default-dev-foo
          namespace: fakeNs
        spec:
          rules:
            - availablePolicy:
                maxUnavailableValue: 30%
              name: maxUnavailable
          selector:
            matchLabels:
              app.kubernetes.io/name: foo
              app.kubernetes.io/part-of: default
        status: {}
      extensions:
        GVK: apps.kusionstack.io/v1alpha1, Kind=PodTransitionRule
phase: succeeded
createTime: 2024-05-10T16:48:00+08:00
modifiedTime: 2024-05-10T16:48:00+08:00
`
}

func mockReleaseMeta(revision uint64) *releaseMetaData {
	return &releaseMetaData{
		Revision: revision,
		Stack:    "test_stack",
	}
}

func mockReleasesMeta() *releasesMetaData {
	return &releasesMetaData{
		LatestRevision: 3,
		ReleaseMetaDatas: []*releaseMetaData{
			mockReleaseMeta(1),
			mockReleaseMeta(2),
			mockReleaseMeta(3),
		},
	}
}

func TestNewLocalStorage(t *testing.T) {
	testcases := []struct {
		name         string
		success      bool
		path         string
		expectedMeta *releasesMetaData
		deletePath   bool
	}{
		{
			name:         "new local storage with empty directory",
			success:      true,
			path:         "empty_releases",
			expectedMeta: &releasesMetaData{},
			deletePath:   true,
		},
		{
			name:         "new local storage with exist directory",
			success:      true,
			path:         "releases",
			expectedMeta: mockReleasesMeta(),
			deletePath:   false,
		},
		{
			name:         "new local storage failed",
			success:      false,
			path:         "invalid_releases",
			expectedMeta: nil,
			deletePath:   false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewLocalStorage(testDataFolder(tc.path))
			assert.Equal(t, tc.success, err == nil)
			if tc.success {
				expectedMetaContent, _ := yaml.Marshal(tc.expectedMeta)
				metaContent, _ := yaml.Marshal(s.meta)
				assert.Equal(t, string(expectedMetaContent), string(metaContent))
			}
			if tc.deletePath {
				pwd, _ := os.Getwd()
				_ = os.RemoveAll(filepath.Join(pwd, "testdata", tc.path))
			}
		})
	}
}

func TestLocalStorage_Get(t *testing.T) {
	testcases := []struct {
		name            string
		success         bool
		revision        uint64
		expectedRelease *v1.Release
	}{
		{
			name:            "get release successfully",
			success:         true,
			revision:        1,
			expectedRelease: mockRelease(1),
		},
		{
			name:            "get release failed not exist",
			success:         false,
			revision:        4,
			expectedRelease: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewLocalStorage(testDataFolder("releases"))
			assert.NoError(t, err)
			r, err := s.Get(tc.revision)
			assert.Equal(t, tc.success, err == nil)
			if tc.success {
				expectedReleaseContent, _ := yaml.Marshal(tc.expectedRelease)
				releaseContent, _ := yaml.Marshal(r)
				assert.Equal(t, string(expectedReleaseContent), string(releaseContent))
			}
		})
	}
}

func TestLocalStorage_GetRevisions(t *testing.T) {
	testcases := []struct {
		name              string
		expectedRevisions []uint64
	}{
		{
			name:              "get release revisions successfully",
			expectedRevisions: []uint64{1, 2, 3},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewLocalStorage(testDataFolder("releases"))
			assert.NoError(t, err)
			revisions := s.GetRevisions()
			assert.Equal(t, tc.expectedRevisions, revisions)
		})
	}
}

func TestLocalStorage_GetStackBoundRevisions(t *testing.T) {
	testcases := []struct {
		name              string
		stack             string
		expectedRevisions []uint64
	}{
		{
			name:              "get stack bound release revisions successfully",
			stack:             "test_stack",
			expectedRevisions: []uint64{1, 2, 3},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewLocalStorage(testDataFolder("releases"))
			assert.NoError(t, err)
			revisions := s.GetStackBoundRevisions(tc.stack)
			assert.Equal(t, tc.expectedRevisions, revisions)
		})
	}
}

func TestLocalStorage_GetLatestRevision(t *testing.T) {
	testcases := []struct {
		name             string
		expectedRevision uint64
	}{
		{
			name:             "get latest release revision successfully",
			expectedRevision: 3,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewLocalStorage(testDataFolder("releases"))
			assert.NoError(t, err)
			revision := s.GetLatestRevision()
			assert.Equal(t, tc.expectedRevision, revision)
		})
	}
}

func TestLocalStorage_Create(t *testing.T) {
	testcases := []struct {
		name         string
		success      bool
		releasePath  string
		revision     uint64
		expectedMeta *releasesMetaData
		deletePath   bool
	}{
		{
			name:        "create release successfully",
			success:     true,
			releasePath: "empty_releases",
			revision:    1,
			expectedMeta: &releasesMetaData{
				LatestRevision: 1,
				ReleaseMetaDatas: []*releaseMetaData{
					mockReleaseMeta(1),
				},
			},
			deletePath: true,
		},
		{
			name:         "create release failed already exist",
			success:      false,
			releasePath:  "releases",
			revision:     3,
			expectedMeta: nil,
			deletePath:   false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewLocalStorage(testDataFolder(tc.releasePath))
			assert.NoError(t, err)
			err = s.Create(mockRelease(tc.revision))
			assert.Equal(t, tc.success, err == nil)
			if tc.success {
				releaseFile := filepath.Join(testDataFolder(tc.releasePath), fmt.Sprintf("%d%s", tc.revision, yamlSuffix))
				_, err = os.Stat(releaseFile)
				assert.NoError(t, err)
			}
			if tc.deletePath {
				pwd, _ := os.Getwd()
				_ = os.RemoveAll(filepath.Join(pwd, "testdata", tc.releasePath))
			}
		})
	}
}

func TestLocalStorage_Update(t *testing.T) {
	testcases := []struct {
		name     string
		success  bool
		revision uint64
	}{
		{
			name:     "update release successfully",
			success:  true,
			revision: 3,
		},
		{
			name:     "update release failed not exist",
			success:  false,
			revision: 4,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewLocalStorage(testDataFolder("releases"))
			assert.NoError(t, err)
			err = s.Update(mockRelease(tc.revision))
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

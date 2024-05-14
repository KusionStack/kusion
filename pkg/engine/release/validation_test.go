package release

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func mockResource() v1.Resource {
	return v1.Resource{
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
	}
}

func TestValidateRelease(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	testcases := []struct {
		name    string
		success bool
		release *v1.Release
	}{
		{
			name:    "valid release",
			success: true,
			release: &v1.Release{
				Project:      "fake-project",
				Workspace:    "fake-ws",
				Revision:     1,
				Stack:        "fake-stack",
				Spec:         &v1.Spec{Resources: v1.Resources{mockResource()}},
				State:        &v1.State{Resources: v1.Resources{mockResource()}},
				Phase:        v1.ReleasePhaseApplying,
				CreateTime:   time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
				ModifiedTime: time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
			},
		},
		{
			name:    "invalid release empty project",
			success: false,
			release: &v1.Release{
				Project:      "",
				Workspace:    "fake-ws",
				Revision:     1,
				Stack:        "fake-stack",
				Spec:         &v1.Spec{Resources: v1.Resources{mockResource()}},
				State:        &v1.State{Resources: v1.Resources{mockResource()}},
				Phase:        v1.ReleasePhaseApplying,
				CreateTime:   time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
				ModifiedTime: time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
			},
		},
		{
			name:    "invalid release empty workspace",
			success: false,
			release: &v1.Release{
				Project:      "fake-project",
				Workspace:    "",
				Revision:     1,
				Stack:        "fake-stack",
				Spec:         &v1.Spec{Resources: v1.Resources{mockResource()}},
				State:        &v1.State{Resources: v1.Resources{mockResource()}},
				Phase:        v1.ReleasePhaseApplying,
				CreateTime:   time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
				ModifiedTime: time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
			},
		},
		{
			name:    "invalid release empty revision",
			success: false,
			release: &v1.Release{
				Project:      "fake-project",
				Workspace:    "fake-ws",
				Revision:     0,
				Stack:        "fake-stack",
				Spec:         &v1.Spec{Resources: v1.Resources{mockResource()}},
				State:        &v1.State{Resources: v1.Resources{mockResource()}},
				Phase:        v1.ReleasePhaseApplying,
				CreateTime:   time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
				ModifiedTime: time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
			},
		},
		{
			name:    "invalid release empty stack",
			success: false,
			release: &v1.Release{
				Project:      "fake-project",
				Workspace:    "fake-ws",
				Revision:     1,
				Stack:        "",
				Spec:         &v1.Spec{Resources: v1.Resources{mockResource()}},
				State:        &v1.State{Resources: v1.Resources{mockResource()}},
				Phase:        v1.ReleasePhaseApplying,
				CreateTime:   time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
				ModifiedTime: time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
			},
		},
		{
			name:    "invalid release invalid spec",
			success: false,
			release: &v1.Release{
				Project:      "fake-project",
				Workspace:    "fake-ws",
				Revision:     1,
				Stack:        "fake-stack",
				Spec:         nil,
				State:        &v1.State{Resources: v1.Resources{mockResource()}},
				Phase:        v1.ReleasePhaseApplying,
				CreateTime:   time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
				ModifiedTime: time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
			},
		},
		{
			name:    "invalid release invalid state",
			success: false,
			release: &v1.Release{
				Project:      "fake-project",
				Workspace:    "fake-ws",
				Revision:     1,
				Stack:        "fake-stack",
				Spec:         &v1.Spec{Resources: v1.Resources{mockResource()}},
				State:        nil,
				Phase:        v1.ReleasePhaseApplying,
				CreateTime:   time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
				ModifiedTime: time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
			},
		},
		{
			name:    "invalid release empty phase",
			success: false,
			release: &v1.Release{
				Project:      "fake-project",
				Workspace:    "fake-ws",
				Revision:     1,
				Stack:        "fake-stack",
				Spec:         &v1.Spec{Resources: v1.Resources{mockResource()}},
				State:        &v1.State{Resources: v1.Resources{mockResource()}},
				Phase:        "",
				CreateTime:   time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
				ModifiedTime: time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
			},
		},
		{
			name:    "invalid release empty create time",
			success: false,
			release: &v1.Release{
				Project:      "fake-project",
				Workspace:    "fake-ws",
				Revision:     1,
				Stack:        "fake-stack",
				Spec:         &v1.Spec{Resources: v1.Resources{mockResource()}},
				State:        &v1.State{Resources: v1.Resources{mockResource()}},
				Phase:        v1.ReleasePhaseApplying,
				ModifiedTime: time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
			},
		},
		{
			name:    "invalid release empty modified time",
			success: false,
			release: &v1.Release{
				Project:    "fake-project",
				Workspace:  "fake-ws",
				Revision:   1,
				Stack:      "fake-stack",
				Spec:       &v1.Spec{Resources: v1.Resources{mockResource()}},
				State:      &v1.State{Resources: v1.Resources{mockResource()}},
				Phase:      v1.ReleasePhaseApplying,
				CreateTime: time.Date(2024, 5, 10, 16, 48, 0, 0, loc),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRelease(tc.release)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateState(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		state   *v1.State
	}{
		{
			name:    "valid state",
			success: true,
			state: &v1.State{
				Resources: v1.Resources{mockResource()},
			},
		},
		{
			name:    "invalid state nil state",
			success: false,
			state:   nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateState(tc.state)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateSpec(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		spec    *v1.Spec
	}{
		{
			name:    "valid spec",
			success: true,
			spec: &v1.Spec{
				Resources: v1.Resources{mockResource()},
			},
		},
		{
			name:    "invalid spec nil spec",
			success: false,
			spec:    nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSpec(tc.spec)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateResources(t *testing.T) {
	testcases := []struct {
		name      string
		success   bool
		resources v1.Resources
	}{
		{
			name:      "valid resources",
			success:   true,
			resources: v1.Resources{mockResource()},
		},
		{
			name:      "invalid resources duplicate resource key",
			success:   false,
			resources: v1.Resources{mockResource(), mockResource()},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateResources(tc.resources)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

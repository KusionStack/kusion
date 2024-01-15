package inputs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

func TestProvider_SetString(t *testing.T) {
	tests := []struct {
		name          string
		data          string
		expected      *Provider
		expectedError error
	}{
		{
			name: "Valid Provider URL",
			data: "registry.terraform.io/hashicorp/aws/5.0.1",
			expected: &Provider{
				URL:       "registry.terraform.io/hashicorp/aws/5.0.1",
				Host:      "registry.terraform.io",
				Namespace: "hashicorp",
				Name:      "aws",
				Version:   "5.0.1",
			},
			expectedError: nil,
		},
		{
			name:     "Invalid Provider URL",
			data:     "registry.terraform.io/hashicorp/aws/invalid-field/5.0.1",
			expected: nil,
			expectedError: fmt.Errorf("wrong provider url format: %s",
				"registry.terraform.io/hashicorp/aws/invalid-field/5.0.1"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testProvider := &Provider{}
			actualErr := testProvider.SetString(test.data)
			if test.expectedError == nil {
				assert.Equal(t, test.expected, testProvider)
				assert.NoError(t, actualErr)
			} else {
				assert.ErrorContains(t, actualErr, test.expectedError.Error())
			}
		})
	}
}

func TestGetProviderURL(t *testing.T) {
	tests := []struct {
		name        string
		data        *apiv1.ProviderConfig
		expected    string
		expectedErr error
	}{
		{
			name: "Default Hashicorp Registry Provider Config",
			data: &apiv1.ProviderConfig{
				Source:  "hashicorp/aws",
				Version: "5.0.1",
			},
			expected:    "registry.terraform.io/hashicorp/aws/5.0.1",
			expectedErr: nil,
		},
		{
			name: "Customized Registry Provider Config",
			data: &apiv1.ProviderConfig{
				Source:  "registry.customized.io/hashicorp/aws",
				Version: "5.0.1",
			},
			expected:    "registry.customized.io/hashicorp/aws/5.0.1",
			expectedErr: nil,
		},
		{
			name: "Empty Version Provider Config",
			data: &apiv1.ProviderConfig{
				Source: "hashicorp/aws",
			},
			expected:    "",
			expectedErr: fmt.Errorf(errEmptyProviderVersion),
		},
		{
			name: "Invalid Provider Source",
			data: &apiv1.ProviderConfig{
				Source:  "aws",
				Version: "5.0.1",
			},
			expected:    "",
			expectedErr: fmt.Errorf(errInvalidProviderSource, "aws"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, actualErr := GetProviderURL(test.data)
			if test.expectedErr == nil {
				assert.Equal(t, test.expected, actual)
				assert.NoError(t, actualErr)
			} else {
				assert.ErrorContains(t, actualErr, test.expectedErr.Error())
			}
		})
	}
}

func TestGetProviderRegion(t *testing.T) {
	tests := []struct {
		name     string
		data     *apiv1.ProviderConfig
		expected string
	}{
		{
			name: "Valid Provider Config",
			data: &apiv1.ProviderConfig{
				Source:  "hashicorp/aws",
				Version: "5.0.1",
				GenericConfig: apiv1.GenericConfig{
					"region": "us-east-1",
				},
			},
			expected: "us-east-1",
		},
		{
			name: "Empty Provider Region",
			data: &apiv1.ProviderConfig{
				Source:  "hashicorp/aws",
				Version: "5.0.1",
			},
			expected: "",
		},
	}

	for _, test := range tests {
		actual := GetProviderRegion(test.data)
		assert.Equal(t, test.expected, actual)
	}
}

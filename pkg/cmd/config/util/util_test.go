package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetItemFromArgs(t *testing.T) {
	testcases := []struct {
		name         string
		success      bool
		args         []string
		expectedItem string
	}{
		{
			name:         "successfully get item",
			success:      true,
			args:         []string{"backends.current"},
			expectedItem: "backends.current",
		},
		{
			name:         "failed to get item invalid args",
			success:      false,
			args:         []string{},
			expectedItem: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			item, err := GetItemFromArgs(tc.args)
			assert.Equal(t, tc.success, err == nil)
			if tc.success {
				assert.Equal(t, tc.expectedItem, item)
			}
		})
	}
}

func TestGetItemValueFromArgs(t *testing.T) {
	testcases := []struct {
		name          string
		success       bool
		args          []string
		expectedItem  string
		expectedValue string
	}{
		{
			name:          "successfully get item and value",
			success:       true,
			args:          []string{"backends.current", "oss-prod"},
			expectedItem:  "backends.current",
			expectedValue: "oss-prod",
		},
		{
			name:          "failed to get item and value invalid args",
			success:       false,
			args:          []string{"backends.current"},
			expectedItem:  "",
			expectedValue: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			item, value, err := GetItemValueFromArgs(tc.args)
			assert.Equal(t, tc.success, err == nil)
			if tc.success {
				assert.Equal(t, tc.expectedItem, item)
				assert.Equal(t, tc.expectedValue, value)
			}
		})
	}
}

func TestValidateNoArg(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		args    []string
	}{
		{
			name:    "no arg",
			success: true,
			args:    []string{},
		},
		{
			name:    "exist arg",
			success: false,
			args:    []string{"backends.current"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateNoArg(tc.args)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateItem(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		item    string
	}{
		{
			name:    "valid item",
			success: true,
			item:    "backends.current",
		},
		{
			name:    "invalid item empty",
			success: false,
			item:    "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateItem(tc.item)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateValue(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		value   string
	}{
		{
			name:    "valid value",
			success: true,
			value:   "oss-prod",
		},
		{
			name:    "invalid value empty",
			success: false,
			value:   "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateValue(tc.value)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

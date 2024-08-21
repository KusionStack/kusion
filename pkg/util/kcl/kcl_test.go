package kcl

import (
	"testing"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func TestAssembleKCLHealthCheck(t *testing.T) {
	tests := []struct {
		name        string
		hpCode      string
		resource    []byte
		want        string
		expectError bool
	}{
		{
			name:        "Valid input",
			hpCode:      "assert res.a == res.b",
			resource:    []byte("this is resource"),
			want:        "import yaml\n\nres = yaml.decode(\"this is resource\")\nassert res.a == res.b\n",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := assembleKCLHealthCheck(tt.hpCode, tt.resource)
			if (err != nil) != tt.expectError {
				t.Errorf("assembleKCLHealthCheck() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if got != tt.want {
				t.Errorf("assembleKCLHealthCheck() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunKCLHealthCheck(t *testing.T) {
	tests := []struct {
		name        string
		hpCode      string
		resource    []byte
		expectError bool
	}{
		{
			name:        "Valid input",
			hpCode:      "a = \"this is health policy\"",
			resource:    []byte("this is resource"),
			expectError: false,
		},
		{
			name:        "evaluation error",
			hpCode:      "assert res.a.b == 1",
			resource:    []byte("res = 2"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunKCLHealthCheck(tt.hpCode, tt.resource)
			if (err != nil) != tt.expectError {
				t.Errorf("RunKCLHealthCheck() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestConvertKCLCode(t *testing.T) {
	tests := []struct {
		name         string
		healthPolicy any
		want         string
		expectOk     bool
	}{
		{
			name: "Valid KCL code in health policy",
			healthPolicy: map[string]any{
				v1.FieldKCLHealthCheckKCL: "assert res.a == res.b",
			},
			want:     "assert res.a == res.b",
			expectOk: true,
		},
		{
			name: "No KCL code in health policy",
			healthPolicy: map[string]any{
				"other_field": "other_value",
			},
			want:     "",
			expectOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ConvertKCLCode(tt.healthPolicy)
			if got != tt.want {
				t.Errorf("ConvertKCLCode() got = %v, want %v", got, tt.want)
			}
			if ok != tt.expectOk {
				t.Errorf("ConvertKCLCode() ok = %v, expectOk %v", ok, tt.expectOk)
			}
		})
	}
}

func Test_validateKCLHealthCheck(t *testing.T) {
	tests := []struct {
		name             string
		healthPolicyCode string
		wantErr          bool
	}{
		{
			name:             "invalid kcl code",
			healthPolicyCode: "assert res.a = res.b",
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateKCLHealthCheck(tt.healthPolicyCode); (err != nil) != tt.wantErr {
				t.Errorf("validateKCLHealthCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

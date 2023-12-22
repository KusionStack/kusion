package tfops

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"kusionstack.io/kusion/pkg/apis/core/v1"
)

var providerAddr = "registry.terraform.io/hashicorp/local/2.2.3"

func TestConvertTFState(t *testing.T) {
	tests := map[string]struct {
		args StateRepresentation
		want v1.Resource
	}{
		"success": {
			args: StateRepresentation{
				FormatVersion:    "0.2",
				TerraformVersion: "1.0.6",
				Values: &stateValues{
					RootModule: module{
						Resources: []resource{
							{
								Address:       "local_file.test",
								Mode:          "managed",
								Type:          "local_file",
								Name:          "test",
								ProviderName:  "registry.terraform.io/hashicorp/local",
								SchemaVersion: 0,
								AttributeValues: attributeValues{
									"content":              "kusion",
									"directory_permission": "0777",
									"file_permission":      "0777",
									"filename":             "text.txt",
								},
							},
						},
					},
				},
			},
			want: v1.Resource{
				ID:   "test",
				Type: "Terraform",
				Attributes: map[string]interface{}{
					"content":              "kusion",
					"directory_permission": "0777",
					"file_permission":      "0777",
					"filename":             "text.txt",
				},
				Extensions: map[string]interface{}{
					"provider":     "registry.terraform.io/hashicorp/local/2.2.3",
					"resourceType": "local_file",
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			state := ConvertTFState(&tc.args, providerAddr)
			if diff := cmp.Diff(tc.want, state); diff != "" {
				t.Errorf("\nConvertTFStateFailed(...) -want message, +got message: \n%s", diff)
			}
		})
	}
}

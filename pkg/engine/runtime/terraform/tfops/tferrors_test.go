package tfops

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var applyInfos = `{"@level":"info","@message":"Terraform 1.0.6","@module":"terraform.ui","@timestamp":"2022-07-28T17:47:22.522277+08:00","terraform":"1.0.6","type":"version","ui":"0.1.0"}
	{"@level":"error","@message":"Error: Extraneous JSON object property","@module":"terraform.ui","@timestamp":"2022-07-28T17:47:22.898885+08:00","diagnostic":{"severity":"error","summary":"Extraneous JSON object property","detail":"No argument or block type is named \"content!\". Did you mean \"content\"?","range":{"filename":"main.tf.json","start":{"line":1,"column":62,"byte":61},"end":{"line":1,"column":72,"byte":71}},"snippet":{"context":"resource.local_file.test","code":"{\"provider\":{\"local\":null},\"resource\":{\"local_file\":{\"test\":{\"content!\":\"kusion12345\",\"filename\":\"test.txt\"}}},\"terraform\":{\"required_providers\":{\"local\":{\"source\":\"registry.terraform.io/hashicorp/local\",\"version\":\"2.2.3\"}}}}","start_line":1,"highlight_start_offset":61,"highlight_end_offset":71,"values":[]}},"type":"diagnostic"}`

func TestParseTerraformInfo(t *testing.T) {
	type args struct {
		infos []byte
	}
	tests := map[string]struct {
		args           args
		wantErrMessage string
	}{
		"PlanError": {
			args: args{
				infos: []byte(`{"@level":"info","@message":"Terraform 1.0.6","@module":"terraform.ui","@timestamp":"2022-07-27T15:57:25.538747+08:00","terraform":"1.0.6","type":"version","ui":"0.1.0"}
				{"@level":"error","@message":"Error: Extraneous JSON object property","@module":"terraform.ui","@timestamp":"2022-07-27T15:57:25.824426+08:00","diagnostic":{"severity":"error","summary":"Extraneous JSON object property","detail":"No argument or block type is named \"content!\". Did you mean \"content\"?","range":{"filename":"main.tf.json","start":{"line":1,"column":62,"byte":61},"end":{"line":1,"column":72,"byte":71}},"snippet":{"context":"resource.local_file.test","code":"{\"provider\":{\"local\":null},\"resource\":{\"local_file\":{\"test\":{\"content!\":\"kusion12345\",\"filename\":\"test.txt\"}}},\"terraform\":{\"required_providers\":{\"local\":{\"source\":\"registry.terraform.io/hashicorp/local\",\"version\":\"2.2.3\"}}}}","start_line":1,"highlight_start_offset":61,"highlight_end_offset":71,"values":[]}},"type":"diagnostic"}`),
			},
			wantErrMessage: "plan failed: Missing required argument: The argument \"location\" is required, but no definition was found.: File name: main.tf.json\nMissing required argument: The argument \"name\" is required, but no definition was found.: File name: main.tf.json",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			TFInfos, err := parseTerraformInfo(tc.args.infos)
			if err != nil {
				t.Errorf("parseTerraformInfo error: %v", err)
			}
			fmt.Println(TFInfos)
		})
	}
}

func TestTFError(t *testing.T) {
	type args struct {
		infos []byte
	}
	tests := map[string]struct {
		args        args
		wantMessage string
	}{
		"apply": {
			args: args{
				infos: []byte(applyInfos),
			},
			wantMessage: "Extraneous JSON object property: No argument or block type is named \"content!\". Did you mean \"content\"?",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := TFError(tc.args.infos)
			got := ""
			if err != nil {
				got = err.Error()
			}
			if diff := cmp.Diff(tc.wantMessage, got); diff != "" {
				t.Errorf("\nWrapApplyFailed(...): -want message, +got message:\n%s", diff)
			}
		})
	}
}

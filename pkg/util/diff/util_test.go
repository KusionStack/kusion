package diff

import (
	"testing"

	"github.com/gonvenience/ytbx"

	"kusionstack.io/kusion/third_party/dyff"
)

func TestToReportString(t *testing.T) {
	type args struct {
		report dyff.Report
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "t1",
			args: args{
				report: dyff.Report{
					From:  ytbx.InputFile{},
					To:    ytbx.InputFile{},
					Diffs: []dyff.Diff{},
				},
			},
			want:    "\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToReportString(tt.args.report)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToReportString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToReportString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToReport(t *testing.T) {
	type args struct {
		oldData interface{}
		newData interface{}
	}
	tests := []struct {
		name      string
		args      args
		wantDiffs int
		wantErr   bool
	}{
		{
			name: "compare string, 1 diff",
			args: args{
				oldData: "a: foo",
				newData: "a: Foo",
			},
			wantDiffs: 1,
		},
		{
			name: "compare struct type, 2 diff",
			args: args{
				oldData: map[string]interface{}{
					"a": "foo",
					"b": 1,
				},
				newData: map[string]interface{}{
					"a": "Foo",
					"b": 2,
				},
			},
			wantDiffs: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToReport(tt.args.oldData, tt.args.newData)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToReport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got.Diffs) != tt.wantDiffs {
				reportStr, _ := ToReportString(*got)
				t.Errorf("ToReport() got = %v, gotDiffs %v, want %v",
					reportStr, len(got.Diffs), tt.wantDiffs)
			}
		})
	}
}

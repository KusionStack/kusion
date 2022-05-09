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

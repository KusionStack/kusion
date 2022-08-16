package s3

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func TestS3Backend_ConfigSchema(t *testing.T) {
	tests := []struct {
		name string
		want cty.Type
	}{
		{
			name: "t1",
			want: cty.Object(map[string]cty.Type{
				"endpoint":        cty.String,
				"bucket":          cty.String,
				"accessKeyID":     cty.String,
				"accessKeySecret": cty.String,
				"region":          cty.String,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewS3Backend()
			if got := s.ConfigSchema(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("S3Backend.ConfigSchema() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestS3Backend_Configure(t *testing.T) {
	defer monkey.UnpatchAll()
	type args struct {
		config map[string]interface{}
	}
	tests := []struct {
		name string
		args
		wantErr bool
	}{
		{
			name: "t1",
			args: args{
				config: map[string]interface{}{
					"endpoint":        "kusion-s3-endpoint",
					"bucket":          "kusion-s3-bucket",
					"accessKeyID":     "kusion-accesskeyID",
					"accessKeySecret": "kusion-accessKeySecret",
					"region":          "kusion-region",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewS3Backend()
			obj, _ := gocty.ToCtyValue(tt.args.config, s.ConfigSchema())
			if err := s.Configure(obj); (err != nil) != tt.wantErr {
				t.Errorf("S3Backend.Configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

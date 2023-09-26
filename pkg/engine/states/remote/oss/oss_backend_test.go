package oss

import (
	"github.com/bytedance/mockey"
	"reflect"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func TestOssBackend_ConfigSchema(t *testing.T) {
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
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewOssBackend()
			if got := s.ConfigSchema(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OSSBackend.ConfigSchema() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOssBackend_Configure(t *testing.T) {
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
					"endpoint":        "oss-cn-hangzhou.aliyuncs.com",
					"bucket":          "kusion-test",
					"accessKeyID":     "kusion-test",
					"accessKeySecret": "kusion-test",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		mockey.PatchConvey(tt.name, t, func() {
			s := NewOssBackend()
			mockOssNew()
			obj, _ := gocty.ToCtyValue(tt.args.config, s.ConfigSchema())
			if err := s.Configure(obj); (err != nil) != tt.wantErr {
				t.Errorf("OSSBackend.Configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func mockOssNew() {
	mockey.Mock(oss.New).To(func(endpoint, accessKeyID, accessKeySecret string, options ...oss.ClientOption) (*oss.Client, error) {
		return &oss.Client{}, nil
	}).Build()
}

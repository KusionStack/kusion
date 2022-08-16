package http

import (
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func TestHttpBackend_ConfigSchema(t *testing.T) {
	tests := []struct {
		name string
		want cty.Type
	}{
		{
			name: "t1",
			want: cty.Object(map[string]cty.Type{
				"urlPrefix":          cty.String,
				"applyURLFormat":     cty.String,
				"getLatestURLFormat": cty.String,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewHTTPBackend()
			if got := s.ConfigSchema(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HttpBackend.ConfigSchema() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpBackend_Configure(t *testing.T) {
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
					"urlPrefix":          "kusion-url",
					"applyURLFormat":     "/apis/v1/tenants/%s/projects/%s/stacks/%s/clusters/%s/states/",
					"getLatestURLFormat": "/apis/v1/tenants/%s/projects/%s/stacks/%s/clusters/%s/states/",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewHTTPBackend()
			obj, _ := gocty.ToCtyValue(tt.args.config, s.ConfigSchema())
			if err := s.Configure(obj); (err != nil) != tt.wantErr {
				t.Errorf("HttBackend.Configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

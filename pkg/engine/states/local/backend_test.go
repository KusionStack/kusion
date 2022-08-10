package local

import (
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func TestLocalBackend_ConfigSchema(t *testing.T) {
	type fields struct {
		Path string
	}
	tests := []struct {
		name   string
		fields fields
		want   cty.Type
	}{
		{
			name: "t1",
			fields: fields{
				Path: stateFile,
			},
			want: cty.Object(map[string]cty.Type{
				"path": cty.String,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewLocalBackend()
			if got := s.ConfigSchema(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LocalBackend.ConfigSchema() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocalBackend_Configure(t *testing.T) {
	type fields struct {
		Path string
	}
	type args struct {
		config map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "t1",
			fields: fields{
				Path: stateFile,
			},
			wantErr: false,
			args: args{
				config: map[string]interface{}{
					"path": stateFile,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewLocalBackend()
			obj, _ := gocty.ToCtyValue(tt.args.config, s.ConfigSchema())
			if err := s.Configure(obj); (err != nil) != tt.wantErr {
				t.Errorf("LocalBackend.Configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

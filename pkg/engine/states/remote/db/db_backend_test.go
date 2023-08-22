package db

import (
	"database/sql"
	"github.com/bytedance/mockey"
	"reflect"
	"testing"

	"github.com/didi/gendry/manager"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func TestDBBackend_ConfigSchema(t *testing.T) {
	tests := []struct {
		name string
		want cty.Type
	}{
		{
			name: "t1",
			want: cty.Object(map[string]cty.Type{
				"dbName":     cty.String,
				"dbUser":     cty.String,
				"dbPassword": cty.String,
				"dbHost":     cty.String,
				"dbPort":     cty.Number,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDBBackend()
			if got := s.ConfigSchema(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBBackend.ConfigSchema() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBBackend_Configure(t *testing.T) {
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
					"dbName":     "kusion-db",
					"dbUser":     "kusion",
					"dbPassword": "kusion",
					"dbHost":     "kusion-host",
					"dbPort":     3306,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		mockey.PatchConvey(tt.name, t, func() {
			s := NewDBBackend()
			mockDBOpen()
			obj, _ := gocty.ToCtyValue(tt.args.config, s.ConfigSchema())
			if err := s.Configure(obj); (err != nil) != tt.wantErr {
				t.Errorf("DBBackend.Configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func mockDBOpen() {
	mockey.Mock((*manager.Option).Open).To(func(o *manager.Option, ping bool) (*sql.DB, error) {
		return &sql.DB{}, nil
	}).Build()
}

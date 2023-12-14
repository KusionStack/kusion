package mysql

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/bytedance/mockey"
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
				"dbName":   cty.String,
				"user":     cty.String,
				"password": cty.String,
				"host":     cty.String,
				"port":     cty.Number,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMysqlBackend()
			if got := s.ConfigSchema(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MysqlBackend.ConfigSchema() = %v, want %v", got, tt.want)
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
					"dbName":   "kusion-db",
					"user":     "kusion",
					"password": "kusion",
					"host":     "kusion-host",
					"port":     3306,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		mockey.PatchConvey(tt.name, t, func() {
			s := NewMysqlBackend()
			mockDBOpen()
			obj, _ := gocty.ToCtyValue(tt.args.config, s.ConfigSchema())
			if err := s.Configure(obj); (err != nil) != tt.wantErr {
				t.Errorf("MysqlBackend.Configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func mockDBOpen() {
	mockey.Mock((*manager.Option).Open).To(func(o *manager.Option, ping bool) (*sql.DB, error) {
		return &sql.DB{}, nil
	}).Build()
}

//go:build !arm64
// +build !arm64

package states

import (
	"database/sql"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/didi/gendry/manager"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"kusionstack.io/kusion/pkg/engine/dal/mapper"
)

func TestNewDBState(t *testing.T) {
	tests := []struct {
		name string
		want StateStorage
	}{
		{
			name: "t1",
			want: &DBState{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDBState(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDBState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBState_ConfigSchema(t *testing.T) {
	type fields struct {
		DB *sql.DB
	}
	tests := []struct {
		name   string
		fields fields
		want   cty.Type
	}{
		{
			name: "t1",
			fields: fields{
				DB: &sql.DB{},
			},
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
			s := &DBState{
				DB: tt.fields.DB,
			}
			if got := s.ConfigSchema(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBState.ConfigSchema() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBState_do2Bo(t *testing.T) {
	type fields struct {
		DB *sql.DB
	}
	type args struct {
		args *mapper.StateDO
	}
	test := []struct {
		name   string
		fields fields
		args   args
		want   *State
	}{
		{
			name: "t1",
			fields: fields{
				DB: &sql.DB{},
			},
			args: args{
				&mapper.StateDO{
					ID:            1,
					GlobalTenant:  "testTenant",
					Env:           "testEnv",
					Project:       "testProject",
					Version:       1,
					KusionVersion: "test",
					Serial:        1,
					Operator:      "test",
					Resources:     "",
				},
			},
			want: &State{
				ID:            1,
				Project:       "testProject",
				Version:       1,
				KusionVersion: "test",
				Serial:        1,
				Operator:      "test",
				Resources:     nil,
			},
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			if got := do2Bo(tt.args.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("do2Bo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func DBStateSetUp(t *testing.T) *DBState {
	monkey.Patch((*manager.Option).Open, func(o *manager.Option, ping bool) (*sql.DB, error) {
		return &sql.DB{}, nil
	})

	stateDo := &mapper.StateDO{GlobalTenant: "test_global_tenant", Project: "test_project", Env: "test_env"}

	monkey.Patch(mapper.GetOne, func(db *sql.DB, where map[string]interface{}) (*mapper.StateDO, error) {
		return stateDo, nil
	})

	monkey.Patch(mapper.Insert, func(db *sql.DB, data []map[string]interface{}) (int64, error) {
		return 1, nil
	})

	return &DBState{DB: &sql.DB{}}
}

func TestDBState(t *testing.T) {
	defer monkey.UnpatchAll()
	dbState := DBStateSetUp(t)

	config := map[string]interface{}{
		"dbName":     "kusion",
		"dbUser":     "kusion",
		"dbPassword": "kusion",
		"dbHost":     "127.0.0.1",
		"dbPort":     3306,
	}
	ctyValue, _ := gocty.ToCtyValue(config, dbState.ConfigSchema())
	err := dbState.Configure(ctyValue)
	assert.NoError(t, err)

	_, err = dbState.GetLatestState(&StateQuery{Tenant: "test_global_tenant", Stack: "test_env", Project: "test_project"})
	assert.NoError(t, err)

	state := &State{Tenant: "test_global_tenant", Project: "test_project", Stack: "test_env"}
	err = dbState.Apply(state)
	assert.NoError(t, err)

	defer func() {
		if r := recover(); r != "implement me" {
			t.Errorf("Delete() got: %v, want: 'implement me'", r)
		}
	}()
	dbState.Delete("test")
}

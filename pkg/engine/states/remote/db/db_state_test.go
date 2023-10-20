//go:build !arm64
// +build !arm64

package db

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/didi/gendry/manager"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/engine/dal/mapper"
	"kusionstack.io/kusion/pkg/engine/states"
)

func TestNewDBState(t *testing.T) {
	tests := []struct {
		name string
		want states.StateStorage
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

func DBStateSetUp(t *testing.T) *DBState {
	mockey.Mock((*manager.Option).Open).To(func(o *manager.Option, ping bool) (*sql.DB, error) {
		return &sql.DB{}, nil
	}).Build()

	stateDo := &mapper.StateDO{Tenant: "test_global_tenant", Project: "test_project", Stack: "test_env"}

	mockey.Mock(mapper.GetOne).To(func(db *sql.DB, where map[string]interface{}) (*mapper.StateDO, error) {
		return stateDo, nil
	}).Build()

	mockey.Mock(mapper.Insert).To(func(db *sql.DB, data []map[string]interface{}) (int64, error) {
		return 1, nil
	}).Build()

	return &DBState{DB: &sql.DB{}}
}

func TestDBState(t *testing.T) {
	mockey.PatchConvey("test DB state", t, func() {
		dbState := DBStateSetUp(t)

		_, err := dbState.GetLatestState(&states.StateQuery{Tenant: "test_global_tenant", Stack: "test_env", Project: "test_project"})
		assert.NoError(t, err)

		state := &states.State{Tenant: "test_global_tenant", Project: "test_project", Stack: "test_env", KusionVersion: "1.0.3"}
		err = dbState.Apply(state)
		assert.NoError(t, err)

		defer func() {
			if r := recover(); r != "implement me" {
				t.Errorf("Delete() got: %v, want: 'implement me'", r)
			}
		}()
		dbState.Delete("test")
	})
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
		want   *states.State
	}{
		{
			name: "t1",
			fields: fields{
				DB: &sql.DB{},
			},
			args: args{
				&mapper.StateDO{
					ID:            1,
					Tenant:        "testTenant",
					Project:       "testProject",
					Stack:         "testEnv",
					Cluster:       "testCluster",
					Version:       1,
					KusionVersion: "test",
					Serial:        1,
					Operator:      "test",
					Resources:     "",
				},
			},
			want: &states.State{
				ID:            1,
				Tenant:        "testTenant",
				Project:       "testProject",
				Stack:         "testEnv",
				Cluster:       "testCluster",
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

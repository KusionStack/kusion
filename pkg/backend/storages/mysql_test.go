package storages

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/engine/state"
	statestorages "kusionstack.io/kusion/pkg/engine/state/storages"
)

func TestNewMysqlStorage(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *v1.BackendMysqlConfig
	}{
		{
			name:    "new mysql storage successfully",
			success: true,
			config: &v1.BackendMysqlConfig{
				DBName: "kusion",
				User:   "kk",
				Host:   "127.0.0.1",
				Port:   3306,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock gorm db", t, func() {
				mockey.Mock(gorm.Open).Return(&gorm.DB{}, nil).Build()
				_, err := NewMysqlStorage(tc.config)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestMysqlStorage_StateStorage(t *testing.T) {
	testcases := []struct {
		name                      string
		mysqlStorage              *MysqlStorage
		project, stack, workspace string
		stateStorage              state.Storage
	}{
		{
			name: "state storage from mysql",
			mysqlStorage: &MysqlStorage{
				db: &gorm.DB{},
			},
			project:   "wordpress",
			stack:     "dev",
			workspace: "dev",
			stateStorage: statestorages.NewMysqlStorage(
				&gorm.DB{},
				"wordpress",
				"dev",
				"dev",
			),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			stateStorage := tc.mysqlStorage.StateStorage(tc.project, tc.stack, tc.workspace)
			assert.Equal(t, stateStorage, tc.stateStorage)
		})
	}
}

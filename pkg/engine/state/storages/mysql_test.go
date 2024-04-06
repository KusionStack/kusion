package storages

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func mockStateMysqlDO() *StateMysqlDO {
	return &StateMysqlDO{
		Project:   "wordpress",
		Stack:     "dev",
		Workspace: "dev",
		Content:   mockStateContent(),
	}
}

func mockMysqlStorage() *MysqlStorage {
	return &MysqlStorage{db: &gorm.DB{}}
}

func TestMysqlStorage_Get(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		stateDO *StateMysqlDO
		state   *v1.State
	}{
		{
			name:    "get mysql state successfully",
			success: true,
			stateDO: mockStateMysqlDO(),
			state:   mockState(),
		},
		{
			name:    "get empty mysql state successfully",
			success: true,
			stateDO: nil,
			state:   nil,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock mysql get", t, func() {
				mockey.Mock(getStateFromMysql).Return(tc.stateDO, nil).Build()
				state, err := mockMysqlStorage().Get()
				assert.Equal(t, tc.success, err == nil)
				assert.Equal(t, tc.state, state)
			})
		})
	}
}

func TestMysqlStorage_Apply(t *testing.T) {
	testcases := []struct {
		name        string
		success     bool
		lastStateDO *StateMysqlDO
		state       *v1.State
	}{
		{
			name:        "update mysql state successfully",
			success:     true,
			lastStateDO: mockStateMysqlDO(),
			state:       mockState(),
		},
		{
			name:        "create mysql state successfully",
			success:     true,
			lastStateDO: nil,
			state:       mockState(),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock mysql create and update", t, func() {
				mockey.Mock(getStateFromMysql).Return(tc.lastStateDO, nil).Build()
				mockey.Mock(checkStateExistenceInMysql).Return(tc.lastStateDO != nil, nil).Build()
				mockey.Mock(createStateInMysql).Return(nil).Build()
				mockey.Mock(updateStateInMysql).Return(nil).Build()
				err := mockMysqlStorage().Apply(tc.state)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

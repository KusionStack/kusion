package storages

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
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

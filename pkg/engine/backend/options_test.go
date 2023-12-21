package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/apis/core/v1"
	_ "kusionstack.io/kusion/pkg/engine/backend/init"
)

func TestBackendOptions_Validate(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		opts    *BackendOptions
	}{
		{
			name:    "valid backend options",
			success: true,
			opts: &BackendOptions{
				Type: v1.BackendMysql,
				Config: []string{
					"dbName=kusion_db",
					"user=kusion",
					"password=kusion_password",
					"host=127.0.0.1",
					"port=3306",
				},
			},
		},
		{
			name:    "invalid backend options empty type",
			success: false,
			opts: &BackendOptions{
				Type: "",
			},
		},
		{
			name:    "invalid backend options unsupported type",
			success: false,
			opts: &BackendOptions{
				Type: "unsupported type",
			},
		},
		{
			name:    "invalid backend options invalid config format",
			success: false,
			opts: &BackendOptions{
				Type:   "mysql",
				Config: []string{"dbName:kusion_db"},
			},
		},
		{
			name:    "invalid backend options empty config key",
			success: false,
			opts: &BackendOptions{
				Type:   "mysql",
				Config: []string{"=kusion_db"},
			},
		},
		{
			name:    "invalid backend options empty config value",
			success: false,
			opts: &BackendOptions{
				Type:   "mysql",
				Config: []string{"dbName="},
			},
		},
		{
			name:    "invalid backend options unsupported config item",
			success: false,
			opts: &BackendOptions{
				Type:   "mysql",
				Config: []string{"unsupported_dbName=kusion_db"},
			},
		},
		{
			name:    "invalid backend options unsupported local backend config",
			success: false,
			opts: &BackendOptions{
				Type:   "local",
				Config: []string{"path=unsupported_kusion_state.yaml"},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.opts.Validate()
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

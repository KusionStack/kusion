package models

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/infra/util/semaphore"
)

func TestOperation_UpdateSemaphore(t *testing.T) {
	original := os.Getenv(apiv1.MaxConcurrentEnvVar)
	defer os.Setenv(apiv1.MaxConcurrentEnvVar, original)

	testcases := []struct {
		name        string
		env         string
		expectedErr error
		expectedVal int64
	}{
		{
			name:        "Invalid Env Type",
			env:         "not-a-number",
			expectedErr: errors.New("invalid syntax"),
		},
		{
			name:        "Invalid Value (less than 0)",
			env:         "-1",
			expectedErr: errors.New("invalid value"),
		},
		{
			name:        "Invalid Value (larger than 100)",
			env:         "200",
			expectedErr: errors.New("invalid value"),
		},
		{
			name:        "Default Value",
			env:         "",
			expectedErr: nil,
			expectedVal: int64(apiv1.DefaultMaxConcurrent),
		},
		{
			name:        "Customized Value",
			env:         "50",
			expectedErr: nil,
			expectedVal: int64(50),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			op := &Operation{}
			os.Setenv(apiv1.MaxConcurrentEnvVar, tc.env)
			err := op.UpdateSemaphore()
			if tc.expectedErr != nil {
				assert.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, *semaphore.New(tc.expectedVal), *op.Sem)
			}
		})
	}
}

package states

import (
	"testing"

	"kusionstack.io/kusion/pkg/engine/states/local"
)

func TestAddToBackends(t *testing.T) {
	type args struct {
		name    string
		storage func() StateStorage
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "t1",
			args: args{
				name:    "test",
				storage: local.NewFileSystemState,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddToBackends(tt.args.name, tt.args.storage)
		})
	}
}

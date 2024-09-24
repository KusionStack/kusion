package list

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCmd(t *testing.T) {
	tests := []struct {
		name string
		want *cobra.Command
	}{
		{
			name: "Create new command Successfully",
			want: &cobra.Command{Use: "list", Short: "List the applied projects"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the actual NewCmd function
			cmd := NewCmd()

			cmd.Execute()
		})
	}
}

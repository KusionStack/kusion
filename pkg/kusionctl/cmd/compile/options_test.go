package compile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompileOptions_preSet(t *testing.T) {
	type fields struct {
		Settings []string
		Output   string
	}

	want := NewCompileOptions()
	want.Settings = []string{"ci-test/settings.yaml", "kcl.yaml"}
	want.Output = "ci-test/stdout.golden.yaml"

	tests := []struct {
		name   string
		fields fields
		want   *CompileOptions
	}{
		{
			name: "preset-noting",
			fields: fields{
				Settings: []string{"ci-test/settings.yaml", "kcl.yaml"},
				Output:   "ci-test/stdout.golden.yaml",
			},
			want: want,
		},
		{
			name: "preset-everything",
			fields: fields{
				Settings: []string{},
				Output:   "",
			},
			want: want,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewCompileOptions()

			o.Settings = tt.fields.Settings
			o.Output = tt.fields.Output

			o.PreSet(func(cur string) bool {
				return true
			})
			assert.Equal(t, tt.want, o)
		})
	}
}

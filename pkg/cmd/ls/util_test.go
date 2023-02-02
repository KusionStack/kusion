//go:build !arm64
// +build !arm64

package ls

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/AlecAivazis/survey/v2"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/projectstack"
)

func Test_commonSearcher(t *testing.T) {
	type args struct {
		content string
		input   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "input has empty string",
			args: args{
				content: "abc",
				input:   "a b c d",
			},
			want: false,
		},
		{
			name: "input has no empty string",
			args: args{
				content: "abc",
				input:   "abc",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, commonSearcher(tt.args.content, tt.args.input), "commonSearcher(%v, %v)", tt.args.content, tt.args.input)
		})
	}
}

var (
	project = &projectstack.Project{
		ProjectConfiguration: projectstack.ProjectConfiguration{
			Name: "http-echo",
		},
		Path:   "../../projectstack/testdata/appops/http-echo",
		Stacks: []*projectstack.Stack{stack},
	}
	stack = &projectstack.Stack{
		StackConfiguration: projectstack.StackConfiguration{
			Name: "dev",
		},
		Path: "../../projectstack/testdata/appops/http-echo/dev",
	}
)

func Test_promptProjectOrStack(t *testing.T) {
	t.Run("prompt project", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockPromptOutput()

		items := []NameAndPath{project}
		got, err := promptProjectOrStack(items, "project")
		assert.Nil(t, err)
		assert.Equal(t, project, got)
	})

	t.Run("prompt stack", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockPromptOutput()

		items := []NameAndPath{stack}
		got, err := promptProjectOrStack(items, "stack")
		assert.Nil(t, err)
		assert.Equal(t, stack, got)
	})
}

func mockPromptOutput() {
	monkey.Patch(
		survey.AskOne,
		func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			s := p.(*survey.Select)
			reflect.ValueOf(response).Elem().Set(reflect.ValueOf(s.Options[0]))
			return nil
		},
	)
}

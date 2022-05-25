//go:build !arm64
// +build !arm64

package init

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/AlecAivazis/survey/v2"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/scaffold"
)

func patchChooseTemplate() {
	monkey.Patch(chooseTemplate, func(templates []scaffold.Template) (scaffold.Template, error) {
		return templates[0], nil
	})
}

func patchPromptValue() {
	monkey.Patch(promptValue, func(valueType, description, defaultValue string, isValidFn func(value string) error) (string, error) {
		return defaultValue, nil
	})
}

func TestRun(t *testing.T) {
	t.Run("init from official url", func(t *testing.T) {
		patchChooseTemplate()
		patchPromptValue()
		defer monkey.UnpatchAll()

		o := &InitOptions{
			Force: true,
		}
		err := o.Complete(nil)
		assert.Nil(t, err)
		err = o.Run()
		assert.Nil(t, err)
		os.RemoveAll(o.ProjectName)
	})
}

func TestChooseTemplate(t *testing.T) {
	t.Run("choose first", func(t *testing.T) {
		// survey.AskOne
		monkey.Patch(
			survey.AskOne,
			func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
				reflect.ValueOf(response).Elem().Set(reflect.ValueOf("foo1    "))
				return nil
			},
		)
		defer monkey.UnpatchAll()

		// test data
		templates := []scaffold.Template{
			{Name: "foo1"},
			{Name: "foo2"},
		}
		chosen, err := chooseTemplate(templates)
		if err != nil {
			return
		}
		assert.Nil(t, err)
		assert.Equal(t, templates[0], chosen)
	})
}

func TestTemplatesToOptionArrayAndMap(t *testing.T) {
	testTpl := scaffold.Template{
		Dir:          "test",
		Name:         "test",
		Description:  "test",
		Quickstart:   "test",
		StackConfigs: []*scaffold.StackTemplate{},
	}
	type args struct {
		templates []scaffold.Template
	}
	tests := []struct {
		name  string
		args  args
		want  []string
		want1 map[string]scaffold.Template
	}{
		{
			name: "t1",
			args: args{
				templates: []scaffold.Template{
					testTpl,
				},
			},
			want: []string{"test    test"},
			want1: map[string]scaffold.Template{
				"test    test": testTpl,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := templatesToOptionArrayAndMap(tt.args.templates)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("templatesToOptionArrayAndMap() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("templatesToOptionArrayAndMap() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestPromptValue(t *testing.T) {
	valueType := "project-name"
	defaultValue := "foo"
	description := "project name"
	flag := true
	isValidFunc := func(value string) error {
		flag = !flag
		if !flag {
			return fmt.Errorf("invalid value: %s", value)
		} else {
			return nil
		}
	}

	// mock survey.AskOne
	monkey.Patch(survey.AskOne, func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		reflect.ValueOf(response).Elem().Set(reflect.ValueOf(defaultValue))
		return nil
	})
	defer monkey.UnpatchAll()

	t.Run("prompt success", func(t *testing.T) {
		got, err := promptValue(valueType, description, defaultValue, nil)
		assert.Nil(t, err)
		assert.Equal(t, defaultValue, got)
	})
	t.Run("valid failed first and succeed next", func(t *testing.T) {
		got, err := promptValue(valueType, description, defaultValue, isValidFunc)
		assert.Nil(t, err)
		assert.Equal(t, defaultValue, got)
	})
}

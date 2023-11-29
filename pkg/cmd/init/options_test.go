package init

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/scaffold"
)

func patchChooseTemplate() {
	mockey.Mock(chooseTemplate).To(func(templates []scaffold.Template) (scaffold.Template, error) {
		return templates[0], nil
	}).Build()
}

func patchPromptValue() {
	mockey.Mock(promptValue).To(func(valueType, description, defaultValue string, isValidFn func(value string) error) (string, error) {
		return defaultValue, nil
	}).Build()
}

func TestRun(t *testing.T) {
	mockey.PatchConvey("init from official url", t, func() {
		patchChooseTemplate()
		patchPromptValue()

		o := &Options{
			Force: true,
		}
		err := o.Complete(nil)
		assert.Nil(t, err)
		err = o.Run()
		assert.Nil(t, err)
		_ = os.RemoveAll(o.ProjectName)
	})

	mockey.PatchConvey("init templates from official url", t, func() {
		o := &TemplatesOptions{
			Output: jsonOutput,
		}
		err := o.Complete(nil, true)
		assert.Nil(t, err)
		err = o.Run()
		assert.Nil(t, err)
	})
}

func TestChooseTemplate(t *testing.T) {
	mockey.PatchConvey("choose first", t, func() {
		// survey.AskOne
		mockey.Mock(
			survey.AskOne).To(
			func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
				reflect.ValueOf(response).Elem().Set(reflect.ValueOf("foo1    "))
				return nil
			},
		).Build()

		// test data
		templates := []scaffold.Template{
			{Name: "foo1", ProjectTemplate: &scaffold.ProjectTemplate{}},
			{Name: "foo2", ProjectTemplate: &scaffold.ProjectTemplate{}},
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
		Dir:  "test",
		Name: "test",
		ProjectTemplate: &scaffold.ProjectTemplate{
			Description:    "test",
			Quickstart:     "test",
			StackTemplates: []*scaffold.StackTemplate{},
		},
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
		mockey.PatchConvey(tt.name, t, func() {
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
	mockey.Mock(survey.AskOne).To(func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		reflect.ValueOf(response).Elem().Set(reflect.ValueOf(defaultValue))
		return nil
	}).Build()

	mockey.PatchConvey("prompt success", t, func() {
		got, err := promptValue(valueType, description, defaultValue, nil)
		assert.Nil(t, err)
		assert.Equal(t, defaultValue, got)
	})
	mockey.PatchConvey("valid failed first and succeed next", t, func() {
		got, err := promptValue(valueType, description, defaultValue, isValidFunc)
		assert.Nil(t, err)
		assert.Equal(t, defaultValue, got)
	})
}

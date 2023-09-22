//go:build !arm64
// +build !arm64

package scaffold

import (
	"fmt"
	"github.com/bytedance/mockey"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jinzhu/copier"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/gitutil"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

const (
	templateDir  = "internal"
	templateName = "single-stack-sample"
)

var (
	localRoot, err    = filepath.Abs(templateDir)
	localTemplateRepo = TemplateRepository{
		Root:         localRoot,
		SubDirectory: filepath.Join(localRoot, templateName),
	}
	localTemplate = Template{
		Dir:  filepath.Join(localRoot, templateName),
		Name: "single-stack-sample",
		ProjectTemplate: &ProjectTemplate{
			ProjectName: "single-stack-sample",
			Description: "A minimal kusion project of single stack",
			Quickstart:  "kusion apply",
			ProjectFields: []*FieldTemplate{
				{
					Name:        "AppName",
					Description: "The Application Name.",
					Type:        StringField,
					Default:     "nginx",
				},
				{
					Name:        "ProjectName",
					Description: "The Project Name.",
					Type:        StringField,
					Default:     "helloworld",
				},
			},
			StackTemplates: []*StackTemplate{
				{
					Name: "dev",
					Fields: []*FieldTemplate{
						{
							Name:        "Image",
							Description: "The Image Address. Default to 'nginx'.",
							Type:        StringField,
							Default:     "nginx",
						},
					},
				},
			},
		},
	}
)

func TestTemplateRepository_Delete(t *testing.T) {
	t.Run("should delete", func(t *testing.T) {
		tmp, err := os.MkdirTemp("", "tmp-dir-for-test")
		assert.Nil(t, err)
		repo := TemplateRepository{
			Root:         tmp,
			ShouldDelete: true,
		}
		err = repo.Delete()
		assert.Nil(t, err)
	})

	t.Run("", func(t *testing.T) {
		err = localTemplateRepo.Delete()
		assert.Nil(t, err)
	})
}

func TestTemplateRepository_Templates(t *testing.T) {
	t.Run("read from dir", func(t *testing.T) {
		templates, err := localTemplateRepo.Templates()
		assert.Nil(t, err)
		assert.Equal(t, []Template{localTemplate}, templates)
	})

	t.Run("read from subdir", func(t *testing.T) {
		subRepo := TemplateRepository{}
		copier.Copy(&subRepo, &localTemplateRepo)
		subRepo.SubDirectory = localTemplateRepo.Root
		templates, err := subRepo.Templates()
		assert.Nil(t, err)
		assert.Contains(t, templates, localTemplate)
	})
}

func TestLoadTemplate(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    Template
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "deployment",
			args: args{
				path: "internal/single-stack-sample",
			},
			want: localTemplate,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return false
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadTemplate(tt.args.path)
			if !tt.wantErr(t, err, fmt.Sprintf("LoadTemplate(%v)", tt.args.path)) {
				return
			}
			assert.Equalf(t, tt.want, got, "LoadTemplate(%v)", tt.args.path)
		})
	}
}

func Test_retrieveKusionTemplates(t *testing.T) {
	t.Run("retrieve not exists", func(t *testing.T) {
		got, err := retrieveKusionTemplates("mockTemplateName", false)
		assert.NotNil(t, err)
		assert.Empty(t, got)
	})
}

func TestRetrieveTemplates(t *testing.T) {
	t.Run("url templates", func(t *testing.T) {
		_, err := RetrieveTemplates(KusionTemplateGitRepository, true)
		assert.Nil(t, err)
	})

	t.Run("file templates", func(t *testing.T) {
		_, err := RetrieveTemplates(localRoot, false)
		assert.Nil(t, err)
	})

	mockey.PatchConvey("kusion templates", t, func() {
		// gitutil.GitCloneOrPull has internet issue occasionally
		// mock as always succeed
		mockey.Mock(gitutil.GitCloneOrPull).To(func(url string, referenceName plumbing.ReferenceName, path string, shallow bool) error {
			return nil
		}).Build()

		_, err := RetrieveTemplates("", true)
		assert.Nil(t, err)
	})
}

//func Test_cleanupLegacyTemplateDir(t *testing.T) {
//	mockey.PatchConvey("repo not exist", t, func() {
//		mockey.Mock(GetTemplateDir).To(func(subDir string) (string, error) {
//			return os.MkdirTemp("", "tmp-dir-for-test")
//		}).Build()
//
//		err = cleanupLegacyTemplateDir()
//		assert.Nil(t, err)
//	})
//
//	t.Run("clean nothing", func(t *testing.T) {
//		err = cleanupLegacyTemplateDir()
//		assert.Nil(t, err)
//	})
//}

func TestValidateProjectName(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "project name is empty",
			args: args{
				s: "",
			},
			wantErr: func(t assert.TestingT, err2 error, i ...interface{}) bool {
				return true
			},
		},
		{
			name: "project name is not match regexp",
			args: args{
				s: "!@#$%^&*()",
			},
			wantErr: func(t assert.TestingT, err2 error, i ...interface{}) bool {
				return true
			},
		},
		{
			name: "project name is valid",
			args: args{
				s: "abc",
			},
			wantErr: func(t assert.TestingT, err2 error, i ...interface{}) bool {
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, ValidateProjectName(tt.args.s), fmt.Sprintf("ValidateProjectName(%v)", tt.args.s))
		})
	}
}

func TestRenderTemplateFiles(t *testing.T) {
	// dest dir
	tmp, err := os.MkdirTemp("", "tmp-dir-for-test")
	assert.Nil(t, err)
	defer func() {
		err = os.RemoveAll(tmp)
		assert.Nil(t, err)
	}()
	// projectConfigs
	projectConfigs := make(map[string]interface{})
	for _, f := range localTemplate.ProjectFields {
		projectConfigs[f.Name] = f.Default
	}
	// stack2Configs
	stack2Configs := make(map[string]map[string]interface{})
	for _, stack := range localTemplate.StackTemplates {
		configs := make(map[string]interface{})
		for _, f := range stack.Fields {
			configs[f.Name] = f.Default
		}
		stack2Configs[stack.Name] = configs
	}
	err = RenderLocalTemplate(localTemplate.Dir, tmp, true, &TemplateConfig{
		ProjectName:   localTemplate.ProjectName,
		ProjectConfig: projectConfigs,
		StacksConfig:  stack2Configs,
	})
	assert.Nil(t, err)
}

func Test_RenderMemTemplateFiles(t *testing.T) {
	memMapFs := afero.NewMemMapFs()
	prj := "test-proj"
	srcFS, _ := Transfer(GetInternalTemplates())
	err := RenderFSTemplate(
		srcFS, "internal/single-stack-sample",
		memMapFs, prj,
		&TemplateConfig{
			ProjectName: prj,
			ProjectConfig: map[string]interface{}{
				"ServiceName": "frontend-svc",
				"NodePort":    30000,
				"ProjectName": prj,
			},
			StacksConfig: map[string]map[string]interface{}{
				"dev": {
					"Stack":       "dev",
					"Image":       "foo/bar:v1",
					"ClusterName": "minikube",
				},
			},
		})
	assert.Nil(t, err)
	err = WriteToDisk(memMapFs, prj, true)
	defer os.RemoveAll(prj)
	assert.Nil(t, err)
}

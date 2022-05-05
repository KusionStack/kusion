package scaffold

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"bou.ke/monkey"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/gitutil"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

const (
	templateDir  = "internal"
	templateName = "deployment-single-stack"
)

var (
	localRoot, err    = filepath.Abs(templateDir)
	localTemplateRepo = TemplateRepository{
		Root:         localRoot,
		SubDirectory: filepath.Join(localRoot, templateName),
	}
	localTemplate = Template{
		Dir:         filepath.Join(localRoot, templateName),
		Name:        "deployment-single-stack",
		ProjectName: "my-app",
		Description: "A minimal kusion project of single stack",
		Quickstart:  "kusion compile main.k -Y ci-test/settings.yaml",
		CommonConfigs: []*FieldTemplate{
			{
				Name:        "ServiceName",
				Description: "service name",
				Type:        StringField,
				Default:     "frontend-svc",
			},
			{
				Name:        "NodePort",
				Description: "node port",
				Type:        IntField,
				Default:     30000,
			},
			{
				Name:        "ProjectName",
				Description: "project name",
				Type:        StringField,
				Default:     "my-app",
			},
		},
		StackConfigs: []*StackTemplate{
			{
				Name: "dev",
				Fields: []*FieldTemplate{
					{
						Name:        "Stack",
						Description: "stack env. One of dev,test,stable,pre,sim,gray,prod.",
						Type:        StringField,
						Default:     "dev",
					},
					{
						Name:        "Image",
						Description: "The Image Address. Default to 'gcr.io/google-samples/gb-frontend:v4'",
						Type:        StringField,
						Default:     "gcr.io/google-samples/gb-frontend:v4",
					},
					{
						Name:        "ClusterName",
						Description: "The Cluster Name. Default to 'kubernetes-dev'",
						Type:        StringField,
						Default:     "kubernetes-dev",
					},
				},
			},
		},
	}
)

func TestTemplateRepository_Delete(t *testing.T) {
	t.Run("should delete", func(t *testing.T) {
		tmp, err := ioutil.TempDir("", "tmp-dir-for-test")
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
		originalSubDir := localTemplateRepo.SubDirectory
		defer func() {
			localTemplateRepo.SubDirectory = originalSubDir
		}()

		localTemplateRepo.SubDirectory = localTemplateRepo.Root
		templates, err := localTemplateRepo.Templates()
		assert.Nil(t, err)
		assert.Equal(t, []Template{localTemplate}, templates)
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
				path: "internal/deployment-single-stack",
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

	t.Run("kusion templates", func(t *testing.T) {
		defer monkey.UnpatchAll()
		// gitutil.GitCloneOrPull has internet issue occasionally
		// mock as always succeed
		monkey.Patch(gitutil.GitCloneOrPull, func(url string, referenceName plumbing.ReferenceName, path string, shallow bool) error {
			return nil
		})

		_, err := RetrieveTemplates("", true)
		assert.Nil(t, err)
	})
}

func Test_cleanupLegacyTemplateDir(t *testing.T) {
	t.Run("repo not exist", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(GetTemplateDir, func(subDir string) (string, error) {
			return ioutil.TempDir("", "tmp-dir-for-test")
		})

		err = cleanupLegacyTemplateDir()
		assert.Nil(t, err)
	})

	t.Run("clean nothing", func(t *testing.T) {
		err = cleanupLegacyTemplateDir()
		assert.Nil(t, err)
	})
}

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

func TestCopyTemplateFiles(t *testing.T) {
	// dest dir
	tmp, err := ioutil.TempDir("", "tmp-dir-for-test")
	assert.Nil(t, err)
	defer func() {
		err = os.RemoveAll(tmp)
		assert.Nil(t, err)
	}()
	// projectConfigs
	projectConfigs := make(map[string]interface{})
	for _, f := range localTemplate.CommonConfigs {
		projectConfigs[f.Name] = f.Default
	}
	// stack2Configs
	stack2Configs := make(map[string]map[string]interface{})
	for _, stack := range localTemplate.StackConfigs {
		configs := make(map[string]interface{})
		for _, f := range stack.Fields {
			configs[f.Name] = f.Default
		}
		stack2Configs[stack.Name] = configs
	}
	// project name
	projectName := localTemplate.ProjectName
	err = CopyTemplateFiles(localRoot, tmp, true, projectName, projectConfigs, stack2Configs)
	assert.Nil(t, err)
}

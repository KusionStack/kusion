package mod

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/gitutil"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

type InitOptions struct {
	Name        string
	Path        string
	TemplateURL string
}

var example = i18n.T(`# Create a kusion module template in the current directory
  kusion mod init my-module

  # Init a kusion module at the specified Path
  kusion mod init my-module ./modules

  # Init a module from a remote git template repository 
  kusion mod init my-module --template https://github.com/<user>/<repo>`)
var short = i18n.T("Create a kusion module along with common files and directories in the current directory")

const (
	defaultTemplateURL = "https://github.com/KusionStack/kusion-module-scaffolding.git"
	defaultBranch      = "main"
)

// NewCmdInit returns an initialized Command instance for the 'mod init' sub command
func NewCmdInit() *cobra.Command {
	o := &InitOptions{}

	cmd := &cobra.Command{
		Use:     "init [MODULE NAME] [PATH]",
		Short:   short,
		Example: templates.Examples(example),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			util.CheckErr(o.Validate(args))
			util.CheckErr(o.Run())
			return
		},
	}
	cmd.Flags().StringVar(&o.TemplateURL, "template", "", i18n.T("Initialize with specified template"))

	return cmd
}

func (o *InitOptions) Validate(args []string) error {
	// get the module Name
	if len(args) < 1 {
		return fmt.Errorf("module Name is empty")
	}
	o.Name = args[0]

	// get the Path
	if len(args) == 2 {
		o.Path = args[1]
	} else {
		// default to the current directory
		o.Path, _ = os.Getwd()
	}

	// create the module directory if not exists
	fs, err := os.Stat(o.Path)
	if err != nil {
		return fmt.Errorf("failed to create module directory: %w", err)
	} else if !fs.IsDir() {
		return fmt.Errorf("path is not a directory")
	} else if os.IsNotExist(err) {
		if err = os.MkdirAll(o.Path, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create module directory: %v", err)
		}
	}
	return nil
}

func (o *InitOptions) Run() error {
	if o.TemplateURL == "" {
		o.TemplateURL = defaultTemplateURL
	}

	// remove existing directory
	dir := filepath.Join(o.Path, o.Name)
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove existing directory: %v", err)
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	// clone templates repo
	branch := plumbing.NewBranchReferenceName(defaultBranch)
	err := gitutil.GitCloneOrPull(o.TemplateURL, branch, dir, false)
	if err != nil {
		return fmt.Errorf("failed to clone git repo:%s, %w", o.TemplateURL, err)
	}
	fmt.Printf("initialized module %s successfully\n", o.Name)
	return nil
}

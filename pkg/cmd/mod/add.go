package mod

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"
	"kcl-lang.io/kpm/pkg/api"
	pkg "kcl-lang.io/kpm/pkg/package"

	"kusionstack.io/kusion/pkg/cmd/meta"
	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var addExample = i18n.T(`# Add a kusion module to the kcl.mod from the current workspace to use it in AppConfiguration 
  kusion mod add my-module

  # Add a module to the kcl.mod from a specified workspace to use it in AppConfiguration 
  kusion mod add my-module --workspace=dev
`)
var addShort = i18n.T("Add a module from a workspace")

type AddFlag struct {
	meta.MetaFlags
}

func (f *AddFlag) toOption(moduleName string, io genericiooptions.IOStreams) (*AddOptions, error) {
	options, err := f.ToOptions()
	if err != nil {
		return nil, err
	}
	return &AddOptions{
		MetaOptions: *options,
		ModuleName:  moduleName,
		IO:          io,
	}, nil
}

type AddOptions struct {
	meta.MetaOptions
	// ModuleName referenced the module name to be added to the kcl.mod file.
	ModuleName string
	// IOStreams referenced the target IOStreams for this invocation.
	IO genericiooptions.IOStreams
}

func (o *AddOptions) Run() error {
	workspace := o.RefWorkspace
	if workspace == nil {
		return fmt.Errorf("cannot find workspace with empty name")
	}

	m := workspace.Modules[o.ModuleName]
	if m == nil {
		return fmt.Errorf("module: %s not found in the workspace: %s", o.ModuleName, workspace.Name)
	}

	// Add module to kcl.mod file
	stack := o.RefStack
	if stack == nil {
		return fmt.Errorf("cannot find stack with empty name")
	}
	kclPkg, err := api.GetKclPackage(stack.Path)
	if err != nil {
		return err
	}
	dependencies := kclPkg.GetDependenciesInModFile()
	if dependencies == nil {
		dependencies = &pkg.Dependencies{Deps: make(map[string]pkg.Dependency)}
	}

	// path example: oci://ghcr.io/kusionstack/service
	u, err := url.Parse(m.Path)
	if err != nil {
		// at least two parts: host and module name are required
		return fmt.Errorf("invalid module path: %s", m.Path)
	}
	if u.Host == "" || u.Path == "" {
		return fmt.Errorf("invalid module path: %s", m.Path)
	}

	dependencies.Deps[o.ModuleName] = pkg.Dependency{
		Name:     o.ModuleName,
		FullName: o.ModuleName + "_" + m.Version,
		Version:  m.Version,
		Source: pkg.Source{
			Oci: &pkg.Oci{
				Reg:  u.Host,
				Repo: strings.TrimPrefix(u.Path, "/"),
				Tag:  m.Version,
			},
		},
	}

	// Save the dependencies to the kcl.mod file
	err = kclPkg.StoreModFile()
	if err != nil {
		return err
	}
	return nil
}

func NewCmdAdd(io genericiooptions.IOStreams) *cobra.Command {
	f := &AddFlag{
		MetaFlags: *meta.NewMetaFlags(),
	}

	cmd := &cobra.Command{
		Use:     "add MODULE_NAME [--workspace WORKSPACE]",
		Short:   addShort,
		Example: templates.Examples(addExample),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)

			o, err := f.toOption(args[0], io)
			util.CheckErr(err)
			util.CheckErr(o.Run())
			return
		},
	}
	f.MetaFlags.AddFlags(cmd)

	return cmd
}

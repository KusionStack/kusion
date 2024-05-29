package mod

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/cmd/meta"
	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var listExample = i18n.T(`# List kusion modules in the current workspace 
  kusion mod list

  # List modules in a specified workspace  
  kusion mod list --workspace=dev 
`)
var listShort = i18n.T("List kusion modules in a workspace ")

type ListFlag struct {
	meta.MetaFlags
}

func (f *ListFlag) toOption(io genericiooptions.IOStreams) (*ListOption, error) {
	storageBackend, err := f.ParseBackend()
	if err != nil {
		return nil, err
	}

	// Get current workspace from backend
	workspace, err := f.ParseWorkspace(storageBackend)
	if err != nil {
		return nil, err
	}

	return &ListOption{
		Workspace:      workspace,
		StorageBackend: storageBackend,
		IO:             io,
	}, nil
}

type ListOption struct {
	// Workspace referenced the workspace for this CLI invocation.
	Workspace *v1.Workspace
	// StorageBackend referenced the target storage backend for this CLI invocation.
	StorageBackend backend.Backend
	// IOStreams referenced the target IOStreams for this CLI invocation.
	IO genericiooptions.IOStreams
}

func (o *ListOption) Run() error {
	tableHeader := []string{"Name", "Version", "URL"}
	tableData := pterm.TableData{tableHeader}
	workspace := o.Workspace
	if workspace == nil {
		return fmt.Errorf("cannot find workspace with nil")
	}
	modules := workspace.Modules
	for k, module := range modules {
		tableData = append(tableData, []string{k, module.Version, module.Path})
	}
	_ = pterm.DefaultTable.WithHasHeader().
		WithHeaderStyle(&pterm.ThemeDefault.TableHeaderStyle).
		WithLeftAlignment(true).
		WithSeparator("  ").
		WithData(tableData).
		WithWriter(o.IO.Out).
		Render()
	pterm.Println()

	return nil
}

func NewCmdList(io genericiooptions.IOStreams) *cobra.Command {
	f := &ListFlag{
		MetaFlags: *meta.NewMetaFlags(),
	}

	cmd := &cobra.Command{
		Use:     "list [--workspace WORKSPACE]",
		Short:   listShort,
		Example: templates.Examples(listExample),
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)

			o, err := f.toOption(io)
			util.CheckErr(err)
			util.CheckErr(o.Run())
			return
		},
	}
	f.MetaFlags.AddFlags(cmd)

	return cmd
}

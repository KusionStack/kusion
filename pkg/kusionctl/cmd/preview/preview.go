package preview

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/kusionctl/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var (
	previewShort = `Preview a series of resource changes within the stack`

	previewLong = `
		Preview a series of resource changes within the stack.

		Create or update or delete resources according to the KCL files within a stack.
		By default, Kusion will generate an execution plan and present it for your approval before taking any action.`

	previewExample = `
		# Preview with specifying work directory
		kusion preview -w /path/to/workdir

		# Preview with specifying arguments
		kusion preview -D name=test -D age=18

		# Preview with specifying setting file
		kusion preview -Y settings.yaml`
)

func NewCmdPreview() *cobra.Command {
	o := NewPreviewOptions()

	cmd := &cobra.Command{
		Use:     "preview",
		Short:   i18n.T(previewShort),
		Long:    templates.LongDesc(i18n.T(previewLong)),
		Example: templates.Examples(i18n.T(previewExample)),
		RunE: func(_ *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			o.Complete(args)
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
			return
		},
	}

	cmd.Flags().StringVarP(&o.CompileOptions.WorkDir, "workdir", "w", "",
		i18n.T("Specify the work directory"))
	cmd.Flags().StringSliceVarP(&o.CompileOptions.Arguments, "argument", "D", []string{},
		i18n.T("Specify the arguments to preview KCL"))
	cmd.Flags().StringSliceVarP(&o.CompileOptions.Settings, "setting", "Y", []string{},
		i18n.T("Specify the command line setting files"))
	cmd.Flags().StringSliceVarP(&o.CompileOptions.Overrides, "overrides", "O", []string{},
		i18n.T("Specify the configuration override path and value"))
	cmd.Flags().BoolVarP(&o.Yes, "yes", "y", false,
		i18n.T("Show preview only, no details"))
	cmd.Flags().BoolVarP(&o.Detail, "detail", "d", false,
		i18n.T("Automatically show plan details after previewing it"))
	cmd.Flags().BoolVarP(&o.NoStyle, "no-style", "", false,
		i18n.T("no-style sets to RawOutput mode and disables all of styling"))

	return cmd
}

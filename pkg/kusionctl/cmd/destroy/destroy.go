package destroy

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/kusionctl/cmd/util"
)

var (
	destroyShort = `Destroy a configuration stack to resource(s) by work directory`

	destroyLong = `
		Delete resources by resource spec.

		Only KCL files are accepted. Only one type of arguments may be specified: filenames,
		resources and names, or resources and label selector.

		Note that the destroy command does NOT do resource version checks, so if someone submits an
		update to a resource right when you submit a destroy, their update will be lost along with the
		rest of the resource.`

	destroyExample = `
		# Delete the configuration of current stack
		kusion destroy`
)

func NewCmdDestroy() *cobra.Command {
	o := NewDestroyOptions()

	cmd := &cobra.Command{
		Use:     "destroy",
		Short:   i18n.T(destroyShort),
		Long:    templates.LongDesc(i18n.T(destroyLong)),
		Example: templates.Examples(i18n.T(destroyExample)),
		RunE: func(_ *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			o.Complete(args)
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
			return
		},
	}

	cmd.Flags().StringVarP(&o.WorkDir, "workdir", "w", "",
		i18n.T("Specify the work directory"))
	cmd.Flags().StringVarP(&o.Operator, "operator", "", "",
		i18n.T("Specify the operator"))
	cmd.Flags().StringSliceVarP(&o.Arguments, "argument", "D", []string{},
		i18n.T("Specify the arguments for compile KCL"))
	cmd.Flags().StringSliceVarP(&o.Settings, "setting", "Y", []string{},
		i18n.T("Specify the command line setting files"))
	cmd.Flags().StringSliceVarP(&o.Overrides, "overrides", "O", []string{},
		i18n.T("Specify the configuration override path and value"))
	cmd.Flags().BoolVarP(&o.Yes, "yes", "y", false,
		i18n.T("Automatically approve and perform the update after previewing it"))
	cmd.Flags().BoolVarP(&o.Detail, "detail", "d", false,
		i18n.T("Automatically show plan details after previewing it"))

	return cmd
}

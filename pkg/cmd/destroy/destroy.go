package destroy

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
)

func NewCmdDestroy() *cobra.Command {
	var (
		destroyShort = i18n.T(`Delete the specified resources in runtime`)

		destroyLong = i18n.T(`
		Delete resources by resource spec.
	
		Only KCL files are accepted. Only one type of arguments may be specified: filenames,
		resources and names, or resources and label selector.
	
		Note that the destroy command does NOT do resource version checks, so if someone submits an
		update to a resource right when you submit a destroy, their update will be lost along with the
		rest of the resource.`)

		destroyExample = i18n.T(`
		# Delete the configuration of current stack
		kusion destroy`)
	)

	o := NewDestroyOptions()
	cmd := &cobra.Command{
		Use:     "destroy",
		Short:   destroyShort,
		Long:    templates.LongDesc(destroyLong),
		Example: templates.Examples(destroyExample),
		RunE: func(_ *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			o.Complete(args)
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
			return
		},
	}

	o.AddCompileFlags(cmd)
	cmd.Flags().StringVarP(&o.Operator, "operator", "", "",
		i18n.T("Specify the operator"))
	cmd.Flags().BoolVarP(&o.Yes, "yes", "y", false,
		i18n.T("Automatically approve and perform the update after previewing it"))
	cmd.Flags().BoolVarP(&o.Detail, "detail", "d", false,
		i18n.T("Automatically show plan details after previewing it"))
	o.AddBackendFlags(cmd)

	return cmd
}

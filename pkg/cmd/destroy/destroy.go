package destroy

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
)

func NewCmdDestroy() *cobra.Command {
	var (
		destroyShort = i18n.T(`Destroy resources within the stack.`)

		destroyLong = i18n.T(`
		Destroy resources within the stack.

		Please note that the destroy command does NOT perform resource version checks.
		Therefore, if someone submits an update to a resource at the same time you execute a destroy command, 
		their update will be lost along with the rest of the resource.`)

		destroyExample = i18n.T(`
		# Delete resources of current stack
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

	o.AddBuildFlags(cmd)
	cmd.Flags().StringVarP(&o.Operator, "operator", "", "",
		i18n.T("Specify the operator"))
	cmd.Flags().BoolVarP(&o.Yes, "yes", "y", false,
		i18n.T("Automatically approve and perform the update after previewing it"))
	cmd.Flags().BoolVarP(&o.Detail, "detail", "d", false,
		i18n.T("Automatically show plan details after previewing it"))
	o.AddBackendFlags(cmd)

	return cmd
}

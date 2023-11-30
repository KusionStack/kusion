package apply

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmdApply() *cobra.Command {
	var (
		applyShort = i18n.T(`Apply the operational intent of various resources to multiple runtimes`)

		applyLong = i18n.T(`
		Apply a series of resource changes within the stack.
	
		Create, update or delete resources according to the operational intent within a stack.
		By default, Kusion will generate an execution plan and prompt for your approval before performing any actions.
		You can review the plan details and make a decision to proceed with the actions or abort them.`)

		applyExample = i18n.T(`
		# Apply with specified work directory
		kusion apply -w /path/to/workdir
	
		# Apply with specified arguments
		kusion apply -D name=test -D age=18

		# Apply with specified intent file
		kusion apply --intent-file intent.yaml

		# Apply with specifying intent file
		kusion apply --intent-file intent.yaml 
	
		# Skip interactive approval of plan details before applying
		kusion apply --yes
		
		# Apply without output style and color
		kusion apply --no-style=true`)
	)

	o := NewApplyOptions()
	cmd := &cobra.Command{
		Use:     "apply",
		Short:   applyShort,
		Long:    templates.LongDesc(applyLong),
		Example: templates.Examples(applyExample),
		RunE: func(_ *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			o.Complete(args)
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
			return
		},
	}

	o.AddBuildFlags(cmd)
	o.AddPreviewFlags(cmd)
	o.AddBackendFlags(cmd)

	cmd.Flags().BoolVarP(&o.Yes, "yes", "y", false,
		i18n.T("Automatically approve and perform the update after previewing it"))
	cmd.Flags().BoolVarP(&o.DryRun, "dry-run", "", false,
		i18n.T("Preview the execution effect (always successful) without actually applying the changes"))
	cmd.Flags().BoolVarP(&o.Watch, "watch", "", false,
		i18n.T("After creating/updating/deleting the requested object, watch for changes"))

	return cmd
}

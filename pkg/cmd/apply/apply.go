package apply

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmdApply() *cobra.Command {
	var (
		applyShort = i18n.T(`Apply the operation intents of various resources to multiple runtimes`)

		applyLong = i18n.T(`
		Apply a series of resource changes within the stack.
	
		Create or update or delete resources according to the KCL files within a stack.
		By default, Kusion will generate an execution plan and present it for your approval before taking any action.
	
		You can check the plan details and then decide if the actions should be taken or aborted.`)

		applyExample = i18n.T(`
		# Apply with specifying work directory
		kusion apply -w /path/to/workdir
	
		# Apply with specifying arguments
		kusion apply -D name=test -D age=18
	
		# Apply with specifying setting file
		kusion apply -Y settings.yaml

		# Apply with specifying spec file
		kusion apply --spec-file spec.yaml 
	
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

	o.AddCompileFlags(cmd)
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

package apply

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/kusionctl/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var (
	applyShort = `Apply a configuration stack to resource(s) by work directory`

	applyLong = `
		Apply a series of resource changes within the stack.

		Create or update or delete resources according to the KCL files within a stack.
		By default, Kusion will generate an execution plan and present it for your approval before taking any action.

		You can check the plan details and then decide if the actions should be taken or aborted.`

	applyExample = `
		# Apply with specifying work directory
		kusion apply -w /path/to/workdir

		# Apply with specifying arguments
		kusion apply -D name=test -D age=18

		# Apply with specifying setting file
		kusion apply -Y settings.yaml

		# Skip interactive approval of plan details before applying
		kusion apply --yes`
)

func NewCmdApply() *cobra.Command {
	o := NewApplyOptions()

	cmd := &cobra.Command{
		Use:     "apply",
		Short:   i18n.T(applyShort),
		Long:    templates.LongDesc(i18n.T(applyLong)),
		Example: templates.Examples(i18n.T(applyExample)),
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
	cmd.Flags().StringVarP(&o.Operator, "operator", "", "",
		i18n.T("Specify the operator"))
	cmd.Flags().StringSliceVarP(&o.CompileOptions.Arguments, "argument", "D", []string{},
		i18n.T("Specify the arguments to apply KCL"))
	cmd.Flags().StringSliceVarP(&o.CompileOptions.Settings, "setting", "Y", []string{},
		i18n.T("Specify the command line setting files"))
	cmd.Flags().StringSliceVarP(&o.CompileOptions.Overrides, "overrides", "O", []string{},
		i18n.T("Specify the configuration override path and value"))
	cmd.Flags().BoolVarP(&o.Yes, "yes", "y", false,
		i18n.T("Automatically approve and perform the update after previewing it"))
	cmd.Flags().BoolVarP(&o.Detail, "detail", "d", false,
		i18n.T("Automatically show plan details after previewing it"))
	cmd.Flags().BoolVarP(&o.NoStyle, "no-style", "", false,
		i18n.T("no-style sets to RawOutput mode and disables all of styling"))
	cmd.Flags().BoolVarP(&o.DryRun, "dry-run", "", false,
		i18n.T("dry-run to preview the execution effect (always successful) without actually applying the changes"))

	return cmd
}

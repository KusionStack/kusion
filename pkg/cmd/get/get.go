package get

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var (
	getShort = `Display the status of all resource(s) by working directory`

	getLong = `
		View the status of resource(s).
		
		Prints the most important information about the status of resources. You can filter the list using a label
		selector and the --selector flag. If you want to see how resources are drifting, you can use the --show-drift flag.`

	getExample = `
		# Display the status of all resource(s) in the current directory
		kusion get

		# Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)
		// TODO: fix#356 add selector support
		kusion get --selector XXX 
		
		# Display resource(s) drifting condition in the current directory
		kusion get --show-drift`
)

func NewCmdGet() *cobra.Command {
	o := NewGetOptions()

	cmd := &cobra.Command{
		Use:   "get",
		Short: i18n.T(getShort),
		// TODO: fix#356 translate `versionLong` to Chinese
		Long: templates.LongDesc(i18n.T(getLong)),
		// TODO: fix#356 translate `versionLong` to Chinese
		Example: templates.Examples(i18n.T(getExample)),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			o.Complete(args)
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
			return
		},
	}

	o.AddPreviewFlags(cmd)

	cmd.Flags().BoolVar(&o.ShowDrift, "show-drift", false,
		i18n.T("Display resource(s) drift"))
	// TODO: fix#356 add `--selector` flag
	//  cmd.Flags().StringVar(&o.Selector...);
	return cmd
}

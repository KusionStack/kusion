package preview

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmdPreview() *cobra.Command {
	var (
		previewShort = i18n.T(`Preview a series of resource changes within the stack`)

		previewLong = i18n.T(`
		Preview a series of resource changes within the stack.
	
		Create or update or delete resources according to the KCL files within a stack. By default,
		Kusion will generate an execution plan and present it for your approval before taking any action.`)

		previewExample = i18n.T(`
		# Preview with specifying work directory
		kusion preview -w /path/to/workdir
	
		# Preview with specifying arguments
		kusion preview -D name=test -D age=18
	
		# Preview with specifying setting file
		kusion preview -Y settings.yaml

		# Preview with specifying intent file
		kusion preview --intent-file intent.yaml
	
		# Preview with ignored fields
		kusion preview --ignore-fields="metadata.generation,metadata.managedFields
		
		# Preview with json format result
		kusion preview -o json

		# Preview without output style and color
		kusion preview --no-style=true`)
	)

	o := NewPreviewOptions()
	cmd := &cobra.Command{
		Use:     "preview",
		Short:   previewShort,
		Long:    templates.LongDesc(previewLong),
		Example: templates.Examples(previewExample),
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

	return cmd
}

func (o *Options) AddPreviewFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Operator, "operator", "", "",
		i18n.T("Specify the operator"))
	cmd.Flags().BoolVarP(&o.Detail, "detail", "d", false,
		i18n.T("Automatically show plan details with interactive options"))
	cmd.Flags().BoolVarP(&o.All, "all", "a", false,
		i18n.T("Automatically show all plan details, combined use with flag `--detail`"))
	cmd.Flags().BoolVarP(&o.NoStyle, "no-style", "", false,
		i18n.T("no-style sets to RawOutput mode and disables all of styling"))
	cmd.Flags().StringSliceVarP(&o.IgnoreFields, "ignore-fields", "", nil,
		i18n.T("Ignore differences of target fields"))
	cmd.Flags().StringVarP(&o.Output, "output", "o", "",
		i18n.T("Specify the output format"))
	cmd.Flags().StringVarP(&o.IntentFile, "intent-file", "", "",
		i18n.T("Specify the intent file path as input, and the intent file must be located in the working directory or its subdirectories"))
}

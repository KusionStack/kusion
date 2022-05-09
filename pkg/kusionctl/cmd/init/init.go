package init

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/kusionctl/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var (
	initShort = `Initialize KCL file structure and base codes for a new project`

	initLong = `
		kusion init command helps you to generate an scaffolding KCL project.
		Try "kusion init" to simply get a demo project.`

	initExample = `
		# Initialize a new KCL project from internal templates
		kusion init

		# Initialize a new KCL project from external default templates location
		kusion init --online=true

		# Initialize a new KCL project from specific templates location
		kusion init https://github.com/<user>/<repo> --online=true

		# Initialize a new KCL project from local directory
		kusion init /path/to/templates`
)

func NewCmdInit() *cobra.Command {
	o := NewInitOptions()
	cmd := &cobra.Command{
		Use:                   "init",
		Short:                 i18n.T(initShort),
		Long:                  templates.LongDesc(i18n.T(initLong)),
		Example:               templates.Examples(i18n.T(initExample)),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			util.CheckErr(o.Complete(args))
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
			return
		},
	}

	cmd.Flags().StringVar(
		&o.ProjectName, "project-name", "",
		i18n.T("The project name; if not specified, a prompt will request it"))
	cmd.Flags().BoolVar(
		&o.Force, "force", false,
		i18n.T("Forces content to be generated even if it would change existing files"))
	cmd.Flags().BoolVar(
		&o.Online, "online", false,
		i18n.T("Use locally cached templates without making any network requests"))
	cmd.Flags().BoolVar(
		&o.Yes, "yes", false,
		i18n.T("Skip prompts and proceed with default values"))
	return cmd
}

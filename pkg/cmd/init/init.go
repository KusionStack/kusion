package init

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmdInit() *cobra.Command {
	var (
		initShort = i18n.T(`Initialize the scaffolding for a project`)

		initLong = i18n.T(`
		This command initializes the scaffolding for a project, generating a project from an appointed template with correct structure.

		The scaffold templates can be retrieved from local or online. The built-in templates are used by default, self-defined templates are also supported by assigning the template repository path.`)

		initExample = i18n.T(`
		# Initialize a project from internal templates
		kusion init

		# Initialize a project from default online templates
		kusion init --online=true

		# Initialize a project from a specific online template
		kusion init https://github.com/<user>/<repo> --online=true

		# Initialize a project from a specific local template
		kusion init /path/to/templates`)
	)

	o := NewInitOptions()
	cmd := &cobra.Command{
		Use:                   "init",
		Short:                 initShort,
		Long:                  templates.LongDesc(initLong),
		Example:               templates.Examples(initExample),
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
		&o.TemplateName, "template-name", "",
		i18n.T("Initialize with specified template. If not specified, a prompt will request it"))
	cmd.Flags().StringVar(
		&o.ProjectName, "project-name", "",
		i18n.T("Initialize with specified project name. If not specified, a prompt will request it"))
	cmd.Flags().BoolVar(
		&o.Force, "force", false,
		i18n.T("Force generating the scaffolding files, even if it would change the existing files"))
	cmd.PersistentFlags().BoolVar(
		&o.Online, "online", false,
		i18n.T("Use templates from online repository to initialize project, or use locally cached templates"))
	cmd.Flags().BoolVar(
		&o.Yes, "yes", false,
		i18n.T("Skip prompts and proceed with default values"))
	cmd.Flags().StringVar(
		&o.CustomParamsJSON, "custom-params", "",
		i18n.T("Custom params in JSON. If specified, it will be used as the template default value and skip prompts"))

	templatesCmd := newCmdTemplates()
	cmd.AddCommand(templatesCmd)
	return cmd
}

func newCmdTemplates() *cobra.Command {
	var (
		templatesShort = i18n.T(`List templates used to initialize a project`)

		templatesLong = i18n.T(`
		This command gets the descriptions and definitions of the templates which are used to initialize the project scaffolding.`)

		templatesExample = i18n.T(`
		# Get name and description of internal templates
		kusion init templates

		# Get templates from specific templates repository
		kusion init templates https://github.com/<user>/<repo> --online=true`)
	)

	o := NewTemplatesOptions()
	cmd := &cobra.Command{
		Use:                   "templates",
		Short:                 templatesShort,
		Long:                  templates.LongDesc(templatesLong),
		Example:               templates.Examples(templatesExample),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			online, err := cmd.InheritedFlags().GetBool("online")
			if err != nil {
				return err
			}
			util.CheckErr(o.Complete(args, online))
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
			return
		},
	}

	cmd.Flags().StringVarP(
		&o.Output, "output", "o", "",
		i18n.T("Specify the output format of templates. If specified, only support json for now; if not, template name and description is given"))

	return cmd
}

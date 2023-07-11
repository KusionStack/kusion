package init

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var (
	initShort = i18n.T(`Initialize KCL file structure and base codes for a new project`)

	initLong = i18n.T(`
		Kusion init command helps you to generate a scaffolding KCL project.
		Try "kusion init" to simply get a demo project.`)

	initExample = i18n.T(`
		# Initialize a new KCL project from internal templates
		kusion init

		# Initialize a new KCL project from external default templates location
		kusion init --online=true

		# Initialize a new KCL project from specific templates location
		kusion init https://github.com/<user>/<repo> --online=true

		# Initialize a new KCL project from local directory
		kusion init /path/to/templates`)
)

func NewCmdInit() *cobra.Command {
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
		i18n.T("The template name; if not specified, a prompt will request it"))
	cmd.Flags().StringVar(
		&o.ProjectName, "project-name", "",
		i18n.T("The project name; if not specified, a prompt will request it"))
	cmd.Flags().BoolVar(
		&o.Force, "force", false,
		i18n.T("Forces content to be generated even if it would change existing files"))
	cmd.PersistentFlags().BoolVar(
		&o.Online, "online", false,
		i18n.T("Use locally cached templates without making any network requests"))
	cmd.Flags().BoolVar(
		&o.Yes, "yes", false,
		i18n.T("Skip prompts and proceed with default values"))
	cmd.Flags().StringVar(
		&o.CustomParamsJSON, "custom-params", "",
		i18n.T("Custom params in JSON string; if not empty, kusion will skip prompt process and use it as template default value"))

	templatesCmd := newCmdTemplates()
	cmd.AddCommand(templatesCmd)
	return cmd
}

var (
	templatesShort = i18n.T(`List Templates used to initialize a new project`)

	templatesLong = i18n.T(`
		Kusion init templates command helps you get the templates which are used
    to generate a scaffolding KCL project.`)

	templatesExample = i18n.T(`
		# Get name and description of internal templates
		kusion init templates

		# Get templates from specific templates location
		kusion init templates https://github.com/<user>/<repo> --online=true`)
)

func newCmdTemplates() *cobra.Command {
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
		i18n.T("The output format, only support json if specified; if not specified, print template name and description"))

	return cmd
}

// Deprecated: Use Build to generate the Intent instead.
package compile

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmdCompile() *cobra.Command {
	compileShort := i18n.T("Deprecated: Use 'kusion build' to generate the Intent instead")
	compileExample := i18n.T("")

	cmd := &cobra.Command{
		Use:     "compile",
		Short:   compileShort,
		Example: templates.Examples(compileExample),
		Aliases: []string{"cl"},
		RunE: func(_ *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			util.CheckErr(fmt.Errorf("this command is deprecated. Please use `kusion build` to generate the Intent instead"))
			return
		},
	}
	return cmd
}

package mod

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"
	"kusionstack.io/kusion-module-framework/pkg/module/registry"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var (
	loginLong = i18n.T(`
	The login command logins to a oci registry for kusion module artifacts.`)

	loginExample = i18n.T(`
	# Login to a oci registry for kusion module artifacts
	kusion mod login ghcr.io/kusion-module-registry --username username --password password
	
	# Users can also set the username and password in the environment variables
	export KUSION_MODULE_REGISTRY_USERNAME=username
	export KUSION_MODULE_REGISTRY_PASSWORD=password
	kusion mod login ghcr.io/kusion-module-registry`)
)

// LoginModFlags directly reflects the information that CLI is gathering via flags. They will be converted to
// LoginModOptions, which reflects the runtime requirements for the command.
type LoginModFlags struct {
	Username string
	Password string

	genericiooptions.IOStreams
}

// LoginModOptions is a set of options that allows you to login to a oci registry for kusion module artifacts.
// This is the object reflects the runtime needs of a `mod login` command, making the logic itself easy to unit test.
type LoginModOptions struct {
	Host     string
	Username string
	Password string

	genericiooptions.IOStreams
}

// NewLoginModFlags returns a default LoginModFlags.
func NewLoginModFlags(ioStreams genericiooptions.IOStreams) *LoginModFlags {
	return &LoginModFlags{
		IOStreams: ioStreams,
	}
}

// NewCmdLogin returns an initialized Command instance for the `mod login` sub command.
func NewCmdLogin(ioStreams genericiooptions.IOStreams) *cobra.Command {
	flags := NewLoginModFlags(ioStreams)

	cmd := &cobra.Command{
		Use:                   "login KUSION_MODULE_REGISTRY_URL [--username username --password password]",
		DisableFlagsInUseLine: true,
		Short:                 "Login to an oci registry for kusion modules",
		Long:                  templates.LongDesc(loginLong),
		Example:               templates.Examples(loginExample),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			o, err := flags.ToOptions(args, flags.IOStreams)
			defer cmdutil.RecoverErr(&err)
			cmdutil.CheckErr(err)
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
			return
		},
	}

	flags.AddFlags(cmd)

	return cmd
}

// AddFlags registers flags for a cli.
func (flags *LoginModFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&flags.Username, "username", "", "The username of kusion module oci registry.")
	cmd.Flags().StringVar(&flags.Password, "password", "", "The password of kusion module oci registry.")
}

// ToOptions converts from CLI inputs to runtime inputs.
func (flags *LoginModFlags) ToOptions(args []string, ioStreams genericiooptions.IOStreams) (*LoginModOptions, error) {
	if len(args) != 1 {
		return nil, errors.New("less or more than one arg is not accepted")
	}

	var username, password string
	if flags.Username != "" {
		username = flags.Username
	} else {
		username = os.Getenv("KUSION_MODULE_REGISTRY_USERNAME")
	}

	if flags.Password != "" {
		password = flags.Password
	} else {
		password = os.Getenv("KUSION_MODULE_REGISTRY_PASSWORD")
	}

	return &LoginModOptions{
		Host:     args[0],
		Username: username,
		Password: password,
	}, nil
}

// Validate verifies if LoginModOptions is valid and without conflicts.
func (o *LoginModOptions) Validate() error {
	if o.Host == "" {
		return errors.New("empty kusion module registry host")
	}

	if o.Username == "" {
		return errors.New("empty kusion module registry username")
	}

	if o.Password == "" {
		return errors.New("empty kusion module registry password")
	}

	return nil
}

// Run executes the `mod login` command.
func (o *LoginModOptions) Run() (err error) {
	_, err = registry.NewKusionModuleClientWithCredentials(o.Host, o.Username, o.Password)

	return
}

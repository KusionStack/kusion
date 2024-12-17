package mod

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"
	"kcl-lang.io/kpm/pkg/client"
	"kcl-lang.io/kpm/pkg/downloader"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var (
	pullLong = i18n.T(`
	The pull command downloads the kusion modules declared in the kcl.mod file.`)

	pullExample = i18n.T(`
	# Pull the kusion modules declared in the kcl.mod file under current directory
	kusion mod pull
	
	# Pull the kusion modules declared in the kcl.mod file under specified directory
	kusion mod pull /dir/to/kcl.mod
	
	# Pull the kusion modules with private oci registry
	kusion mod pull --host ghcr.io/kusion-module-registry --username username --password password
	
	# Or users can also pull the kusion modules in private oci registry with environment variables
	export KUSION_MODULE_HOST=ghcr.io/kusion-module-registry
	export KUSION_MODULE_USERNAME=username
	export KUSION_MODULE_PASSWORD=password
	kusion mod pull /dir/to/kcl.mod`)
)

// PullModFlags directly reflects the information that CLI is gathering via flags. They will be converted to
// PullModOptions, which reflects the runtime requirements for the command.
//
// This structure reduces the transformation to wiring and makes the logic itself easy to unit test.
type PullModFlags struct {
	Host     string
	Username string
	Password string

	genericiooptions.IOStreams
}

// PullModOptions is a set of options that allows you to push module. This is the object reflects the
// runtime needs of a `mod pull` command, making the logic itself easy to unit test.
type PullModOptions struct {
	Dir      string
	Host     string
	Username string
	Password string

	genericiooptions.IOStreams
}

// NewPullModFlags returns a default PullModFlags.
func NewPullModFlags(ioStreams genericiooptions.IOStreams) *PullModFlags {
	return &PullModFlags{
		IOStreams: ioStreams,
	}
}

// NewCmdPull returns an initialized Command instance for the `mod pull` sub command.
func NewCmdPull(ioStreams genericiooptions.IOStreams) *cobra.Command {
	flags := NewPullModFlags(ioStreams)

	cmd := &cobra.Command{
		Use:                   "pull KCL_MOD_FILE_PATH [--host kusion-module-oci-registry --username username --password password]",
		DisableFlagsInUseLine: true,
		Short:                 "Pull kusion modules",
		Long:                  templates.LongDesc(pullLong),
		Example:               templates.Examples(pullExample),
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
func (flags *PullModFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&flags.Host, "host", "", "The host of kusion module oci registry.")
	cmd.Flags().StringVar(&flags.Username, "username", "", "The username of kusion module oci registry.")
	cmd.Flags().StringVar(&flags.Password, "password", "", "The password of kusion module oci registry.")
}

// ToOptions converts from CLI inputs to runtime inputs.
func (flags *PullModFlags) ToOptions(args []string, ioStreams genericiooptions.IOStreams) (*PullModOptions, error) {
	if len(args) > 1 {
		return nil, errors.New("more than one args are not accepted")
	}

	var dir string
	var err error
	if len(args) > 0 {
		dir = args[0]
	} else {
		dir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get the working directory: %v", err)
		}
	}

	var host, username, password string
	if flags.Host != "" {
		host = flags.Host
	} else {
		host = os.Getenv("KUSION_MODULE_HOST")
	}

	if flags.Username != "" {
		username = flags.Username
	} else {
		username = os.Getenv("KUSION_MODULE_USERNAME")
	}

	if flags.Password != "" {
		password = flags.Password
	} else {
		password = os.Getenv("KUSION_MODULE_PASSWORD")
	}

	return &PullModOptions{
		Dir:      dir,
		Host:     host,
		Username: username,
		Password: password,
	}, nil
}

// Validate verifies if PullModOptions is valid and without conflicts.
func (o *PullModOptions) Validate() error {
	if _, err := os.Stat(filepath.Join([]string{o.Dir, "kcl.mod"}...)); err != nil {
		return fmt.Errorf("no kcl.mod file found at path %s", o.Dir)
	}

	return nil
}

// Run executes the `mod pull` command.
func (o *PullModOptions) Run() (err error) {
	cacheDir := filepath.Join(o.Dir, ".kclvm")
	defer func(path string) {
		if err != nil {
			return
		}
		err = os.RemoveAll(path)
	}(cacheDir)

	cli, err := client.NewKpmClient()
	if err != nil {
		return
	}
	cli.DepDownloader = downloader.NewOciDownloader(runtime.GOOS + "/" + runtime.GOARCH)

	// Login to the private oci registry.
	if o.Host != "" && o.Username != "" && o.Password != "" {
		if err = cli.LoginOci(o.Host, o.Username, o.Password); err != nil {
			return
		}
	}

	// Download the kusion modules with kpm sdk.
	kclPkg, err := cli.LoadPkgFromPath(o.Dir)
	if err != nil {
		return
	}

	_, _, err = cli.InitGraphAndDownloadDeps(kclPkg)
	return
}

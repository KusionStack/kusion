package rel

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/cmd/meta"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var (
	unlockShort = i18n.T("Unlock the latest release file of the current stack")

	unlockLong = i18n.T(`
	Unlock the latest release file of the current stack. 

	The phase of the latest release file of the current stack in the current or a specified workspace
	will be set to 'failed' if it was in the stages of 'generating', 'previewing', 'applying' or 'destroying'. 

	Please note that using the 'kusion release unlock' command may cause unexpected concurrent read-write
	issues with release files, so please use it with caution. 
	`)

	unlockExample = i18n.T(`# Unlock the latest release file of the current stack in the current workspace. 
	kusion release unlock

	# Unlock the latest release file of the current stack in a specified workspace. 
	kusion release unlock --workspace=dev
`)
)

// UnlockFlags reflects the information that CLI is gathering via flags,
// which will be converted into UnlockOptions.
type UnlockFlags struct {
	MetaFlags *meta.MetaFlags
}

// UnlockOptions defines the configuration parameters for the `kusion release unlock` command.
type UnlockOptions struct {
	*meta.MetaOptions
}

// NewUnlockFlags returns a default UnlockFlags.
func NewUnlockFlags(streams genericiooptions.IOStreams) *UnlockFlags {
	return &UnlockFlags{
		MetaFlags: meta.NewMetaFlags(),
	}
}

// NewCmdUnlock creates the `kusion release unlock` command.
func NewCmdUnlock(streams genericiooptions.IOStreams) *cobra.Command {
	flags := NewUnlockFlags(streams)

	cmd := &cobra.Command{
		Use:     "unlock",
		Short:   unlockShort,
		Long:    templates.LongDesc(unlockLong),
		Example: templates.Examples(unlockExample),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			o, err := flags.ToOptions()
			defer cmdutil.RecoverErr(&err)
			cmdutil.CheckErr(err)
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run())

			return
		},
	}

	flags.AddFlags(cmd)

	return cmd
}

// AddFlags registers flags for the CLI.
func (f *UnlockFlags) AddFlags(cmd *cobra.Command) {
	f.MetaFlags.AddFlags(cmd)
}

// ToOptions converts from CLI inputs to runtime inputs.
func (f *UnlockFlags) ToOptions() (*UnlockOptions, error) {
	metaOpts, err := f.MetaFlags.ToOptions()
	if err != nil {
		return nil, err
	}

	o := &UnlockOptions{
		MetaOptions: metaOpts,
	}

	return o, nil
}

// Validate verifies if UnlockOptions are valid and without conflicts.
func (o *UnlockOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageErrorf(cmd, "Unexpected args: %v", args)
	}

	return nil
}

// Run executes the `kusion release unlock` command.
func (o *UnlockOptions) Run() error {
	// Get the storage backend of the release.
	storage, err := o.Backend.ReleaseStorage(o.RefProject.Name, o.RefWorkspace.Name)
	if err != nil {
		return err
	}

	// Get the latest release.
	r, err := release.GetLatestRelease(storage)
	if err != nil {
		return err
	}
	if r == nil {
		fmt.Printf("No release file found for project: %s, workspace: %s\n",
			o.RefProject.Name, o.RefWorkspace.Name)
		return nil
	}

	// Update the phase to 'failed', if it was not succeeded or failed.
	if r.Phase != v1.ReleasePhaseSucceeded && r.Phase != v1.ReleasePhaseFailed {
		r.Phase = v1.ReleasePhaseFailed

		if err := storage.Update(r); err != nil {
			return err
		}

		fmt.Printf("Successfully update release phase to Failed, project: %s, workspace: %s, revision: %d\n",
			r.Project, r.Workspace, r.Revision)

		return nil
	}

	fmt.Printf("No need to update the release phase, current phase: %s\n", r.Phase)
	return nil
}

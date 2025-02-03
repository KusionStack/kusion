// Copyright 2024 KusionStack Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apply

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"kusionstack.io/kusion/pkg/engine/apply"
	applystate "kusionstack.io/kusion/pkg/engine/apply/state"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/cmd/generate"
	"kusionstack.io/kusion/pkg/cmd/preview"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/util/i18n"
	"kusionstack.io/kusion/pkg/util/pretty"
	"kusionstack.io/kusion/pkg/util/terminal"
)

var (
	applyLong = i18n.T(`
		Apply a series of resource changes within the stack.
	
		Create, update or delete resources according to the operational intent within a stack.
		By default, Kusion will generate an execution preview and prompt for your approval before performing any actions.
		You can review the preview details and make a decision to proceed with the actions or abort them.`)

	applyExample = i18n.T(`
		# Apply with specified work directory
		kusion apply -w /path/to/workdir

		# Apply with specified arguments
		kusion apply -D name=test -D age=18
	
		# Apply with specifying spec file
		kusion apply --spec-file spec.yaml

		# Skip interactive approval of preview details before applying
		kusion apply --yes
		
		# Apply without output style and color
		kusion apply --no-style=true
		
		# Apply without watching the resource changes and waiting for reconciliation
		kusion apply --watch=false

		# Apply with the specified timeout duration for kusion apply command, measured in second(s)
		kusion apply --timeout=120

		# Apply with localhost port forwarding
		kusion apply --port-forward=8080`)
)

// ApplyFlags directly reflect the information that CLI is gathering via flags. They will be converted to
// ApplyOptions, which reflect the runtime requirements for the command.
//
// This structure reduces the transformation to wiring and makes the logic itself easy to unit test.
type ApplyFlags struct {
	*preview.PreviewFlags

	Yes         bool
	DryRun      bool
	Watch       bool
	Timeout     int
	PortForward int

	genericiooptions.IOStreams
}

// NewApplyFlags returns a default ApplyFlags
func NewApplyFlags(ui *terminal.UI, streams genericiooptions.IOStreams) *ApplyFlags {
	return &ApplyFlags{
		PreviewFlags: preview.NewPreviewFlags(ui, streams),
		IOStreams:    streams,
	}
}

// NewCmdApply creates the `apply` command.
func NewCmdApply(ui *terminal.UI, ioStreams genericiooptions.IOStreams) *cobra.Command {
	flags := NewApplyFlags(ui, ioStreams)

	cmd := &cobra.Command{
		Use:     "apply",
		Short:   "Apply the operational intent of various resources to multiple runtimes",
		Long:    templates.LongDesc(applyLong),
		Example: templates.Examples(applyExample),
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

// AddFlags registers flags for a cli.
func (f *ApplyFlags) AddFlags(cmd *cobra.Command) {
	// bind flag structs
	f.PreviewFlags.AddFlags(cmd)

	cmd.Flags().BoolVarP(&f.Yes, "yes", "y", false, i18n.T("Automatically approve and perform the update after previewing it"))
	cmd.Flags().BoolVarP(&f.DryRun, "dry-run", "", false, i18n.T("Preview the execution effect (always successful) without actually applying the changes"))
	cmd.Flags().BoolVarP(&f.Watch, "watch", "", true, i18n.T("After creating/updating/deleting the requested object, watch for changes"))
	cmd.Flags().IntVarP(&f.Timeout, "timeout", "", 0, i18n.T("The timeout duration for kusion apply command, measured in second(s)"))
	cmd.Flags().IntVarP(&f.PortForward, "port-forward", "", 0, i18n.T("Forward the specified port from local to service"))
}

// ToOptions converts from CLI inputs to runtime inputs.
func (f *ApplyFlags) ToOptions() (*ApplyOptions, error) {
	// Convert preview options
	previewOptions, err := f.PreviewFlags.ToOptions()
	if err != nil {
		return nil, err
	}

	o := &ApplyOptions{
		PreviewOptions: previewOptions,
		Yes:            f.Yes,
		DryRun:         f.DryRun,
		Watch:          f.Watch,
		Timeout:        f.Timeout,
		PortForward:    f.PortForward,
		IOStreams:      f.IOStreams,
	}

	return o, nil
}

// Validate verifies if ApplyOptions are valid and without conflicts.
func (o *ApplyOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageErrorf(cmd, "Unexpected args: %v", args)
	}

	if o.PortForward < 0 || o.PortForward > 65535 {
		return cmdutil.UsageErrorf(cmd, "Invalid port number to forward: %d, must be between 1 and 65535", o.PortForward)
	}

	if o.SpecFile != "" {
		absSF, _ := filepath.Abs(o.SpecFile)
		fi, err := os.Stat(absSF)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("spec file not exist: %s", absSF)
			}
		}

		if fi.IsDir() || !fi.Mode().IsRegular() {
			return fmt.Errorf("spec file must be a regular file: %s", absSF)
		}
		absWD, _ := filepath.Abs(o.RefStack.Path)

		// calculate the relative path between absWD and absSF,
		// if absSF is not located in the directory or subdirectory specified by absWD,
		// an error will be returned.
		rel, err := filepath.Rel(absWD, absSF)
		if err != nil {
			return err
		}
		if rel[:3] == ".."+string(filepath.Separator) {
			return fmt.Errorf("the spec file must be located in the working directory or its subdirectories of the stack")
		}
	}

	return nil
}

func (o *ApplyOptions) prepareApply() (state *applystate.State, err error) {
	defer cmdutil.RecoverErr(&err)

	// init apply state
	state = &applystate.State{
		Metadata: &applystate.Metadata{
			Project:   o.RefProject.Name,
			Workspace: o.RefWorkspace.Name,
			Stack:     o.RefStack.Name,
		},
		RelLock:     &sync.Mutex{},
		PortForward: o.PortForward,
		DryRun:      o.DryRun,
		Watch:       o.Watch,
		Ls:          &applystate.LineSummary{},
	}

	// create release
	state.ReleaseStorage, err = o.Backend.ReleaseStorage(o.RefProject.Name, o.RefWorkspace.Name)
	if err != nil {
		return
	}

	state.TargetRel, err = release.NewApplyRelease(state.ReleaseStorage, o.RefProject.Name, o.RefStack.Name, o.RefWorkspace.Name)
	if err != nil {
		return
	}

	if !o.DryRun {
		if err = state.CreateStorageRelease(state.TargetRel); err != nil {
			return
		}
	}

	// build parameters
	parameters := make(map[string]string)
	for _, value := range o.PreviewOptions.Values {
		parts := strings.SplitN(value, "=", 2)
		parameters[parts[0]] = parts[1]
	}

	// generate Spec
	var spec *apiv1.Spec
	if o.SpecFile != "" {
		spec, err = generate.SpecFromFile(o.SpecFile)
	} else {
		spec, err = generate.GenerateSpecWithSpinner(o.RefProject, o.RefStack, o.RefWorkspace, parameters, o.UI, o.NoStyle)
	}
	if err != nil {
		return
	}

	// return immediately if no resource found in stack
	if spec == nil || len(spec.Resources) == 0 {
		fmt.Println(pretty.GreenBold("\nNo resource found in this stack."))
		return state, nil
	}

	// prepare target rel done
	state.TargetRel.Spec = spec
	return
}

// Run executes the `apply` command.
func (o *ApplyOptions) Run() (err error) {
	// prepare apply
	applyState, err := o.prepareApply()

	if err != nil && applyState != nil {
		updateErr := applyState.UpdateReleasePhaseFailed()
		if updateErr != nil {
			err = errors.Join(err, updateErr)
		}
	}

	if err != nil {
		return
	}

	// apply action
	err = apply.Apply(o, applyState)
	return
}

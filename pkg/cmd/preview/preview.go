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

package preview

import (
	"encoding/json"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/cmd/generate"
	"kusionstack.io/kusion/pkg/cmd/meta"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform"
	"kusionstack.io/kusion/pkg/engine/state"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util/i18n"
	"kusionstack.io/kusion/pkg/util/pretty"
)

var (
	previewLong = i18n.T(`
		Preview a series of resource changes within the stack.
	
		Create, update or delete resources according to the intent described in the stack. By default,
		Kusion will generate an execution preview and present it for your approval before taking any action.`)

	previewExample = i18n.T(`
		# Preview with specified work directory
		kusion preview -w /path/to/workdir
	
		# Preview with specified arguments
		kusion preview -D name=test -D age=18

		# Preview with ignored fields
		kusion preview --ignore-fields="metadata.generation,metadata.managedFields
		
		# Preview with json format result
		kusion preview -o json

		# Preview without output style and color
		kusion preview --no-style=true`)
)

const jsonOutput = "json"

// PreviewFlags directly reflect the information that CLI is gathering via flags. They will be converted to
// PreviewOptions, which reflect the runtime requirements for the command.
//
// This structure reduces the transformation to wiring and makes the logic itself easy to unit test.
type PreviewFlags struct {
	MetaFlags *meta.MetaFlags

	Operator     string
	Detail       bool
	All          bool
	NoStyle      bool
	Output       string
	IgnoreFields []string

	genericiooptions.IOStreams
}

// PreviewOptions defines flags and other configuration parameters for the `preview` command.
type PreviewOptions struct {
	*meta.MetaOptions

	Operator     string
	Detail       bool
	All          bool
	NoStyle      bool
	Output       string
	IgnoreFields []string

	genericiooptions.IOStreams
}

// NewPreviewFlags returns a default PreviewFlags
func NewPreviewFlags(streams genericiooptions.IOStreams) *PreviewFlags {
	return &PreviewFlags{
		MetaFlags: meta.NewMetaFlags(),
		IOStreams: streams,
	}
}

// NewCmdPreview creates the `preview` command.
func NewCmdPreview(ioStreams genericiooptions.IOStreams) *cobra.Command {
	flags := NewPreviewFlags(ioStreams)

	cmd := &cobra.Command{
		Use:     "preview",
		Short:   "Preview a series of resource changes within the stack",
		Long:    templates.LongDesc(previewLong),
		Example: templates.Examples(previewExample),
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
func (f *PreviewFlags) AddFlags(cmd *cobra.Command) {
	// bind flag structs
	f.MetaFlags.AddFlags(cmd)

	cmd.Flags().StringVarP(&f.Operator, "operator", "", f.Operator, i18n.T("Specify the operator"))
	cmd.Flags().BoolVarP(&f.Detail, "detail", "d", true, i18n.T("Automatically show preview details with interactive options"))
	cmd.Flags().BoolVarP(&f.All, "all", "a", false, i18n.T("Automatically show all preview details, combined use with flag `--detail`"))
	cmd.Flags().BoolVarP(&f.NoStyle, "no-style", "", false, i18n.T("no-style sets to RawOutput mode and disables all of styling"))
	cmd.Flags().StringSliceVarP(&f.IgnoreFields, "ignore-fields", "", f.IgnoreFields, i18n.T("Ignore differences of target fields"))
	cmd.Flags().StringVarP(&f.Output, "output", "o", f.Output, i18n.T("Specify the output format"))
}

// ToOptions converts from CLI inputs to runtime inputs.
func (f *PreviewFlags) ToOptions() (*PreviewOptions, error) {
	// Convert meta options
	metaOptions, err := f.MetaFlags.ToOptions()
	if err != nil {
		return nil, err
	}

	o := &PreviewOptions{
		MetaOptions:  metaOptions,
		Operator:     f.Operator,
		Detail:       f.Detail,
		All:          f.All,
		NoStyle:      f.NoStyle,
		Output:       f.Output,
		IgnoreFields: f.IgnoreFields,
		IOStreams:    f.IOStreams,
	}

	return o, nil
}

// Validate verifies if PreviewOptions are valid and without conflicts.
func (o *PreviewOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageErrorf(cmd, "Unexpected args: %v", args)
	}

	return nil
}

// Run executes the `preview` command.
func (o *PreviewOptions) Run() error {
	// set no style
	if o.NoStyle || o.Output == jsonOutput {
		pterm.DisableStyling()
		pterm.DisableColor()
	}

	// Generate spec
	spec, err := generate.GenerateSpecWithSpinner(o.RefProject, o.RefStack, o.RefWorkspace, nil, o.NoStyle)
	if err != nil {
		return err
	}

	// return immediately if no resource found in stack
	if spec == nil || len(spec.Resources) == 0 {
		if o.Output != jsonOutput {
			fmt.Println(pretty.GreenBold("\nNo resource found in this stack."))
		}
		return nil
	}

	// compute changes for preview
	storage := o.StorageBackend.StateStorage(o.RefProject.Name, o.RefStack.Name, o.RefWorkspace.Name)
	changes, err := Preview(o, storage, spec, o.RefProject, o.RefStack)
	if err != nil {
		return err
	}

	if o.Output == jsonOutput {
		var previewChanges []byte
		previewChanges, err = json.Marshal(changes)
		if err != nil {
			return fmt.Errorf("json marshal preview changes failed as %w", err)
		}
		fmt.Println(string(previewChanges))
		return nil
	}

	if changes.AllUnChange() {
		fmt.Println("All resources are reconciled. No diff found")
		return nil
	}

	// summary preview table
	changes.Summary(o.IOStreams.Out)

	// detail detection
	if o.Detail {
		for {
			var target string
			target, err = changes.PromptDetails()
			if err != nil {
				return err
			}
			if target == "" { // Cancel option
				break
			}
			changes.OutputDiff(target)
		}
	}
	return nil
}

// The Preview function calculates the upcoming actions of each resource
// through the execution Kusion Engine, and you can customize the
// runtime of engine and the state storage through `runtime` and
// `storage` parameters.
//
// Example:
//
//	o := NewPreviewOptions()
//	stateStorage := &states.FileSystemState{
//	    Path: filepath.Join(o.WorkDir, states.KusionState)
//	}
//	kubernetesRuntime, err := runtime.NewKubernetesRuntime()
//	if err != nil {
//	    return err
//	}
//
//	changes, err := Preview(o, kubernetesRuntime, stateStorage,
//	    planResources, project, stack, os.Stdout)
//	if err != nil {
//	    return err
//	}
func Preview(
	opts *PreviewOptions,
	storage state.Storage,
	planResources *apiv1.Spec,
	project *apiv1.Project,
	stack *apiv1.Stack,
) (*models.Changes, error) {
	log.Info("Start compute preview changes ...")

	// check and install terraform executable binary for
	// resources with the type of Terraform.
	tfInstaller := terraform.CLIInstaller{
		Intent: planResources,
	}
	if err := tfInstaller.CheckAndInstall(); err != nil {
		return nil, err
	}

	// construct the preview operation
	pc := &operation.PreviewOperation{
		Operation: models.Operation{
			OperationType: models.ApplyPreview,
			Stack:         stack,
			StateStorage:  storage,
			IgnoreFields:  opts.IgnoreFields,
			ChangeOrder:   &models.ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*models.ChangeStep{}},
		},
	}

	log.Info("Start call pc.Preview() ...")

	// parse cluster in arguments
	rsp, s := pc.Preview(&operation.PreviewRequest{
		Request: models.Request{
			Project:  project,
			Stack:    stack,
			Operator: opts.Operator,
			Intent:   planResources,
		},
	})
	if v1.IsErr(s) {
		return nil, fmt.Errorf("preview failed.\n%s", s.String())
	}

	return models.NewChanges(project, stack, rsp.Order), nil
}

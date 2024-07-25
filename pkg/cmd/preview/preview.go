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
	"os"
	"path/filepath"
	"strings"

	"github.com/liu-hm19/pterm"
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
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util/diff"
	"kusionstack.io/kusion/pkg/util/i18n"
	"kusionstack.io/kusion/pkg/util/pretty"
	"kusionstack.io/kusion/pkg/util/terminal"
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

		# Preview with specifying spec file
		kusion preview --spec-file spec.yaml

		# Preview with ignored fields
		kusion preview --ignore-fields="metadata.generation,metadata.managedFields"
		
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

	Detail       bool
	All          bool
	NoStyle      bool
	Output       string
	SpecFile     string
	IgnoreFields []string
	Values       []string

	UI *terminal.UI

	genericiooptions.IOStreams
}

// PreviewOptions defines flags and other configuration parameters for the `preview` command.
type PreviewOptions struct {
	*meta.MetaOptions

	Detail       bool
	All          bool
	NoStyle      bool
	Output       string
	SpecFile     string
	IgnoreFields []string
	Values       []string

	UI *terminal.UI

	genericiooptions.IOStreams
}

// NewPreviewFlags returns a default PreviewFlags
func NewPreviewFlags(ui *terminal.UI, streams genericiooptions.IOStreams) *PreviewFlags {
	return &PreviewFlags{
		MetaFlags: meta.NewMetaFlags(),
		UI:        ui,
		IOStreams: streams,
	}
}

// NewCmdPreview creates the `preview` command.
func NewCmdPreview(ui *terminal.UI, ioStreams genericiooptions.IOStreams) *cobra.Command {
	flags := NewPreviewFlags(ui, ioStreams)

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

	cmd.Flags().BoolVarP(&f.Detail, "detail", "d", true, i18n.T("Automatically show preview details with interactive options"))
	cmd.Flags().BoolVarP(&f.All, "all", "a", false, i18n.T("Automatically show all preview details, combined use with flag `--detail`"))
	cmd.Flags().BoolVarP(&f.NoStyle, "no-style", "", false, i18n.T("no-style sets to RawOutput mode and disables all of styling"))
	cmd.Flags().StringSliceVarP(&f.IgnoreFields, "ignore-fields", "", f.IgnoreFields, i18n.T("Ignore differences of target fields"))
	cmd.Flags().StringVarP(&f.Output, "output", "o", f.Output, i18n.T("Specify the output format"))
	cmd.Flags().StringArrayVarP(&f.Values, "argument", "D", []string{}, i18n.T("Specify arguments on the command line"))
	cmd.Flags().StringVarP(&f.SpecFile, "spec-file", "", "", i18n.T("Specify the spec file path as input, and the spec file must be located in the working directory or its subdirectories"))
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
		Detail:       f.Detail,
		All:          f.All,
		NoStyle:      f.NoStyle,
		Output:       f.Output,
		SpecFile:     f.SpecFile,
		IgnoreFields: f.IgnoreFields,
		UI:           f.UI,
		IOStreams:    f.IOStreams,
		Values:       f.Values,
	}

	return o, nil
}

// Validate verifies if PreviewOptions are valid and without conflicts.
func (o *PreviewOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageErrorf(cmd, "Unexpected args: %v", args)
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

// Run executes the `preview` command.
func (o *PreviewOptions) Run() error {
	// set no style
	if o.NoStyle || o.Output == jsonOutput {
		pterm.DisableStyling()
	}

	// build parameters
	parameters := make(map[string]string)
	for _, value := range o.Values {
		parts := strings.SplitN(value, "=", 2)
		parameters[parts[0]] = parts[1]
	}

	// Generate spec
	var spec *apiv1.Spec
	var err error
	if o.SpecFile != "" {
		spec, err = generate.SpecFromFile(o.SpecFile)
	} else {
		spec, err = generate.GenerateSpecWithSpinner(o.RefProject, o.RefStack, o.RefWorkspace, parameters, o.UI, o.NoStyle)
	}
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

	// compute state
	storage, err := o.Backend.ReleaseStorage(o.RefProject.Name, o.RefWorkspace.Name)
	if err != nil {
		return err
	}
	state, err := release.GetLatestState(storage)
	if err != nil {
		return err
	}
	if state == nil {
		state = &apiv1.State{}
	}

	// compute changes for preview
	changes, err := Preview(o, storage, spec, state, o.RefProject, o.RefStack)
	if err != nil {
		return err
	}

	if o.Output == jsonOutput {
		var previewChanges []byte

		// Mask sensitive data before printing the preview changes.
		for _, v := range changes.ChangeSteps {
			maskedFrom, maskedTo := diff.MaskSensitiveData(v.From, v.To)
			v.From = maskedFrom
			v.To = maskedTo
		}

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
	changes.Summary(o.IOStreams.Out, o.NoStyle)

	// detail detection
	if o.Detail {
		if o.All {
			changes.OutputDiff("all")
		} else {
			for {
				var target string
				target, err = changes.PromptDetails(o.UI)
				if err != nil {
					return err
				}
				if target == "" { // Cancel option
					break
				}
				changes.OutputDiff(target)
			}
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
//	o := newPreviewOptions()
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
	storage release.Storage,
	planResources *apiv1.Spec,
	priorResources *apiv1.State,
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
			OperationType:  models.ApplyPreview,
			Stack:          stack,
			ReleaseStorage: storage,
			IgnoreFields:   opts.IgnoreFields,
			ChangeOrder:    &models.ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*models.ChangeStep{}},
		},
	}

	log.Info("Start call pc.Preview() ...")

	// parse cluster in arguments
	rsp, s := pc.Preview(&operation.PreviewRequest{
		Request: models.Request{
			Project: project,
			Stack:   stack,
		},
		Spec:  planResources,
		State: priorResources,
	})
	if v1.IsErr(s) {
		return nil, fmt.Errorf("preview failed.\n%s", s.String())
	}

	return models.NewChanges(project, stack, rsp.Order), nil
}

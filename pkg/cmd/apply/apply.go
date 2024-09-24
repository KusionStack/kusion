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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/liu-hm19/pterm"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	"gopkg.in/yaml.v3"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/cmd/generate"
	"kusionstack.io/kusion/pkg/cmd/preview"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/printers"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/engine/resource/graph"
	"kusionstack.io/kusion/pkg/engine/runtime"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util/i18n"
	"kusionstack.io/kusion/pkg/util/kcl"
	"kusionstack.io/kusion/pkg/util/pretty"
	"kusionstack.io/kusion/pkg/util/signal"
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

// To handle the release phase update when panic occurs.
// Fixme: adopt a more centralized approach to manage the release update before exiting, instead of
// scattering them across different go-routines.
var (
	rel            *apiv1.Release
	gph            *apiv1.Graph
	relLock        = &sync.Mutex{}
	releaseCreated = false
	releaseStorage release.Storage
	portForwarded  = false
)

var errExit = errors.New("receive SIGTERM or SIGINT, exit cmd")

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

// ApplyOptions defines flags and other configuration parameters for the `apply` command.
type ApplyOptions struct {
	*preview.PreviewOptions

	SpecFile    string
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
		SpecFile:       f.SpecFile,
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

// Run executes the `apply` command.
func (o *ApplyOptions) Run() (err error) {
	// update release to succeeded or failed
	defer func() {
		if !releaseCreated {
			return
		}
		if err != nil {
			release.UpdateReleasePhase(rel, apiv1.ReleasePhaseFailed, relLock)
			// Join the errors if update apply release failed.
			err = errors.Join([]error{err, release.UpdateApplyRelease(releaseStorage, rel, o.DryRun, relLock)}...)
		} else {
			release.UpdateReleasePhase(rel, apiv1.ReleasePhaseSucceeded, relLock)
			err = release.UpdateApplyRelease(releaseStorage, rel, o.DryRun, relLock)
		}
	}()

	// set no style
	if o.NoStyle {
		pterm.DisableStyling()
	}

	// create release
	releaseStorage, err = o.Backend.ReleaseStorage(o.RefProject.Name, o.RefWorkspace.Name)
	if err != nil {
		return
	}
	rel, err = release.NewApplyRelease(releaseStorage, o.RefProject.Name, o.RefStack.Name, o.RefWorkspace.Name)
	if err != nil {
		return
	}
	if !o.DryRun {
		if err = releaseStorage.Create(rel); err != nil {
			return
		}
		releaseCreated = true
	}

	// Prepare for the timeout timer.
	// Fixme: adopt a more centralized approach to manage the gracefully exit interrupted by
	// the SIGINT or SIGTERM, instead of scattering them across different go-routines.
	var timer <-chan time.Time
	errCh := make(chan error, 1)
	defer close(errCh)

	// Wait for the SIGTERM or SIGINT.
	go func() {
		stopCh := signal.SetupSignalHandler()
		<-stopCh
		errCh <- errExit
	}()

	go func() {
		errCh <- o.run(rel, releaseStorage)
	}()

	// Check whether the kusion apply command has timed out.
	if o.Timeout > 0 {
		timer = time.After(time.Second * time.Duration(o.Timeout))
		select {
		case err = <-errCh:
			if errors.Is(err, errExit) && portForwarded {
				return nil
			}
			return err
		case <-timer:
			err = fmt.Errorf("failed to execute kusion apply as: timeout for %d seconds", o.Timeout)
			if !releaseCreated {
				return
			}
			release.UpdateReleasePhase(rel, apiv1.ReleasePhaseFailed, relLock)
			err = errors.Join([]error{err, release.UpdateApplyRelease(releaseStorage, rel, o.DryRun, relLock)}...)
			return err
		}
	} else {
		err = <-errCh
		if errors.Is(err, errExit) && portForwarded {
			return nil
		}
	}

	return err
}

// run executes the apply cmd after the release is created.
func (o *ApplyOptions) run(rel *apiv1.Release, releaseStorage release.Storage) (err error) {
	defer func() {
		if !releaseCreated {
			return
		}
		if err != nil {
			release.UpdateReleasePhase(rel, apiv1.ReleasePhaseFailed, relLock)
			err = errors.Join([]error{err, release.UpdateApplyRelease(releaseStorage, rel, o.DryRun, relLock)}...)
		}
	}()

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
		return nil
	}

	// update release phase to previewing
	rel.Spec = spec
	release.UpdateReleasePhase(rel, apiv1.ReleasePhasePreviewing, relLock)
	if err = release.UpdateApplyRelease(releaseStorage, rel, o.DryRun, relLock); err != nil {
		return
	}

	// compute changes for preview
	changes, err := preview.Preview(o.PreviewOptions, releaseStorage, rel.Spec, rel.State, o.RefProject, o.RefStack)
	if err != nil {
		return
	}

	if allUnChange(changes) {
		fmt.Println("All resources are reconciled. No diff found")
		return nil
	}

	// summary preview table
	changes.Summary(o.IOStreams.Out, o.NoStyle)

	// detail detection
	if o.Detail && o.All {
		changes.OutputDiff("all")
		if !o.Yes {
			return nil
		}
	}

	// prompt
	if !o.Yes {
		for {
			var input string
			input, err = prompt(o.UI)
			if err != nil {
				return err
			}
			if input == "yes" {
				break
			} else if input == "details" {
				var target string
				target, err = changes.PromptDetails(o.UI)
				if err != nil {
					return err
				}
				changes.OutputDiff(target)
			} else {
				fmt.Println("Operation apply canceled")
				return nil
			}
		}
	}

	// update release phase to applying
	release.UpdateReleasePhase(rel, apiv1.ReleasePhaseApplying, relLock)
	if err = release.UpdateApplyRelease(releaseStorage, rel, o.DryRun, relLock); err != nil {
		return
	}

	// Get graph storage directory, create if not exist
	graphStorage, err := o.Backend.GraphStorage(o.RefProject.Name, o.RefWorkspace.Name)
	if err != nil {
		return err
	}

	// Try to get existing graph, use the graph if exists
	if graphStorage.CheckGraphStorageExistence() {
		gph, err = graphStorage.Get()
		if err != nil {
			return err
		}
		err = graph.ValidateGraph(gph)
		if err != nil {
			return err
		}
		// Put new resources from the generated spec to graph
		gph, err = graph.GenerateGraph(spec.Resources, gph)
	} else {
		// Create a new graph to be used globally if no graph is stored in the storage
		gph = &apiv1.Graph{
			Project:   o.RefProject.Name,
			Workspace: o.RefWorkspace.Name,
		}
		gph, err = graph.GenerateGraph(spec.Resources, gph)
	}
	if err != nil {
		return err
	}

	// start applying
	fmt.Printf("\nStart applying diffs ...\n")

	// NOTE: release should be updated in the process of apply, so as to avoid the problem
	// of being unable to update after being terminated by SIGINT or SIGTERM.
	_, err = Apply(o, releaseStorage, rel, gph, changes)
	if err != nil {
		return
	}

	// if dry run, print the hint
	if o.DryRun {
		fmt.Printf("\nNOTE: Currently running in the --dry-run mode, the above configuration does not really take effect\n")
		return nil
	}

	if o.PortForward > 0 {
		fmt.Printf("\nStart port-forwarding ...\n")
		portForwarded = true
		if err = PortForward(o, rel.Spec); err != nil {
			return
		}
	}

	return
}

// The Apply function will apply the resources changes through the execution kusion engine.
// You can customize the runtime of engine and the release releaseStorage through `runtime` and `releaseStorage` parameters.
func Apply(
	o *ApplyOptions,
	releaseStorage release.Storage,
	rel *apiv1.Release,
	gph *apiv1.Graph,
	changes *models.Changes,
) (*apiv1.Release, error) {
	var err error
	// Update the release before exit.
	defer func() {
		if p := recover(); p != nil {
			cmdutil.RecoverErr(&err)
			log.Error(err)
		}
		if err != nil {
			if !releaseCreated {
				return
			}
			release.UpdateReleasePhase(rel, apiv1.ReleasePhaseFailed, relLock)
			err = errors.Join([]error{err, release.UpdateApplyRelease(releaseStorage, rel, o.DryRun, relLock)}...)
		}

		// Update graph and write to storage if not dry run.
		if !o.DryRun {
			// Use resources in the state to get resource Cloud ID.
			for _, resource := range rel.State.Resources {
				// Get information of each of the resources
				info, err := graph.GetResourceInfo(&resource)
				if err != nil {
					return
				}
				// Update information of each of the resources.
				graphResource := graph.FindGraphResourceByID(gph.Resources, resource.ID)
				if graphResource != nil {
					graphResource.CloudResourceID = info.CloudResourceID
					graphResource.Type = info.ResourceType
					graphResource.Name = info.ResourceName
				}
			}
			// Get the directory to store the graph.
			graphStorage, err := o.Backend.GraphStorage(o.RefProject.Name, o.RefWorkspace.Name)
			if err != nil {
				return
			}

			// Update graph if exists, otherwise create a new graph file.
			if graphStorage.CheckGraphStorageExistence() {
				// No need to store resource index
				graph.RemoveResourceIndex(gph)
				err := graphStorage.Update(gph)
				if err != nil {
					return
				}
			} else {
				graph.RemoveResourceIndex(gph)
				err := graphStorage.Create(gph)
				if err != nil {
					return
				}
			}
		}
	}()

	// construct the apply operation
	ac := &operation.ApplyOperation{
		Operation: models.Operation{
			Stack:          changes.Stack(),
			ReleaseStorage: releaseStorage,
			MsgCh:          make(chan models.Message),
			IgnoreFields:   o.IgnoreFields,
		},
	}

	// Init a watch channel with a sufficient buffer when it is necessary to perform watching.
	if o.Watch && !o.DryRun {
		ac.WatchCh = make(chan string, 100)
	}

	// line summary
	var ls lineSummary
	// Get the multi printer from UI option.
	multi := o.UI.MultiPrinter
	// Max length of resource ID for progressbar width.
	maxLen := 0

	// Prepare the writer to print the operation progress and results.
	changesWriterMap := make(map[string]*pterm.SpinnerPrinter)
	for _, key := range changes.Values() {
		// Get the maximum length of the resource ID.
		if len(key.ID) > maxLen {
			maxLen = len(key.ID)
		}
		// Init a spinner printer for the resource to print the apply status.
		changesWriterMap[key.ID], err = o.UI.SpinnerPrinter.
			WithWriter(multi.NewWriter()).
			Start(fmt.Sprintf("Pending %s", pterm.Bold.Sprint(key.ID)))
		if err != nil {
			return nil, fmt.Errorf("failed to init change step spinner printer: %v", err)
		}
	}

	// Init a writer for progressbar.
	pbWriter := multi.NewWriter()
	// progress bar, print dag walk detail
	progressbar, err := o.UI.ProgressbarPrinter.
		WithTotal(len(changes.StepKeys)).
		WithWriter(pbWriter).
		WithRemoveWhenDone().
		WithShowCount(false).
		WithMaxWidth(maxLen + 32).
		Start()
	if err != nil {
		return nil, err
	}

	// The writer below is for operation error printing.
	errWriter := multi.NewWriter()

	multi.WithUpdateDelay(time.Millisecond * 100)
	multi.Start()
	defer multi.Stop()

	// wait msgCh close
	var wg sync.WaitGroup
	// receive msg and print detail
	go PrintApplyDetails(
		ac,
		&err,
		&errWriter,
		&wg,
		changes,
		changesWriterMap,
		progressbar,
		&ls,
		o.DryRun,
		o.Watch,
		gph.Resources,
	)

	watchErrCh := make(chan error)
	// Apply while watching the resources.
	if o.Watch && !o.DryRun {
		Watch(
			ac,
			changes,
			&err,
			o.DryRun,
			watchErrCh,
			multi,
			changesWriterMap,
			gph,
		)
	}

	var updatedRel *apiv1.Release
	if o.DryRun {
		for _, r := range rel.Spec.Resources {
			ac.MsgCh <- models.Message{
				ResourceID: r.ResourceKey(),
				OpResult:   models.Success,
				OpErr:      nil,
			}
		}
		close(ac.MsgCh)
	} else {
		// parse cluster in arguments
		rsp, st := ac.Apply(&operation.ApplyRequest{
			Request: models.Request{
				Project: changes.Project(),
				Stack:   changes.Stack(),
			},
			Release: rel,
			Graph:   gph,
		})
		if v1.IsErr(st) {
			errWriter.(*bytes.Buffer).Reset()
			err = fmt.Errorf("apply failed, status:\n%v", st)
			return nil, err
		}
		// Update the release with that in the apply response if not dryrun.
		updatedRel = rsp.Release
		*rel = *updatedRel
		gph = rsp.Graph
	}

	// wait for msgCh closed
	wg.Wait()
	// Wait for watchWg closed if need to perform watching.
	if o.Watch && !o.DryRun {
		shouldBreak := false
		for !shouldBreak {
			select {
			case watchErr := <-watchErrCh:
				if watchErr != nil {
					return nil, watchErr
				}
				shouldBreak = true
			default:
				continue
			}
		}
	}

	// print summary
	pterm.Fprintln(pbWriter, fmt.Sprintf("\nApply complete! Resources: %d created, %d updated, %d deleted.", ls.created, ls.updated, ls.deleted))
	return updatedRel, nil
}

// PrintApplyDetails function will receive the messages of the apply operation and print the details.
// Fixme: abstract the input variables into a struct.
func PrintApplyDetails(
	ac *operation.ApplyOperation,
	err *error,
	errWriter *io.Writer,
	wg *sync.WaitGroup,
	changes *models.Changes,
	changesWriterMap map[string]*pterm.SpinnerPrinter,
	progressbar *pterm.ProgressbarPrinter,
	ls *lineSummary,
	dryRun bool,
	watch bool,
	gphResources *apiv1.GraphResources,
) {
	defer func() {
		if p := recover(); p != nil {
			cmdutil.RecoverErr(err)
			log.Error(*err)
		}
		if *err != nil {
			if !releaseCreated {
				return
			}
			release.UpdateReleasePhase(rel, apiv1.ReleasePhaseFailed, relLock)
			*err = errors.Join([]error{*err, release.UpdateApplyRelease(releaseStorage, rel, dryRun, relLock)}...)
		}
		(*errWriter).(*bytes.Buffer).Reset()
	}()
	wg.Add(1)

	for {
		select {
		// Get operation results from the message channel.
		case msg, ok := <-ac.MsgCh:
			if !ok {
				wg.Done()
				return
			}
			changeStep := changes.Get(msg.ResourceID)

			// Update the progressbar and spinner printer according to the operation result.
			switch msg.OpResult {
			case models.Success, models.Skip:
				var title string
				if changeStep.Action == models.UnChanged {
					title = fmt.Sprintf("Skipped %s", pterm.Bold.Sprint(changeStep.ID))
					changesWriterMap[msg.ResourceID].Success(title)
				} else {
					if watch && !dryRun {
						title = fmt.Sprintf("%s %s",
							changeStep.Action.Ing(),
							pterm.Bold.Sprint(changeStep.ID),
						)
						changesWriterMap[msg.ResourceID].UpdateText(title)
					} else {
						changesWriterMap[msg.ResourceID].Success(fmt.Sprintf("Succeeded %s", pterm.Bold.Sprint(msg.ResourceID)))
					}
				}

				// Update resource status
				if !dryRun && changeStep.Action != models.UnChanged {
					gphResource := graph.FindGraphResourceByID(gphResources, msg.ResourceID)
					if gphResource != nil {
						// Delete resource from the graph if it's deleted during apply
						if changeStep.Action == models.Delete {
							graph.RemoveResource(gph, gphResource)
						} else {
							gphResource.Status = apiv1.ApplySucceed
						}
					}
				}

				progressbar.Increment()
				ls.Count(changeStep.Action)
			case models.Failed:
				title := fmt.Sprintf("Failed %s", pterm.Bold.Sprint(changeStep.ID))
				changesWriterMap[msg.ResourceID].Fail(title)
				errStr := pretty.ErrorT.Sprintf("apply %s failed as: %s\n", msg.ResourceID, msg.OpErr.Error())
				pterm.Fprintln(*errWriter, errStr)
				if !dryRun {
					// Update resource status, in case anything like update fail happened
					gphResource := graph.FindGraphResourceByID(gphResources, msg.ResourceID)
					if gphResource != nil {
						gphResource.Status = apiv1.ApplyFail
					}
				}
			default:
				title := fmt.Sprintf("%s %s",
					changeStep.Action.Ing(),
					pterm.Bold.Sprint(changeStep.ID),
				)
				changesWriterMap[msg.ResourceID].UpdateText(title)
			}
		}
	}
}

// Watch function will watch the changed Kubernetes and Terraform resources.
// Fixme: abstract the input variables into a struct.
func Watch(
	ac *operation.ApplyOperation,
	changes *models.Changes,
	err *error,
	dryRun bool,
	watchErrCh chan error,
	multi *pterm.MultiPrinter,
	changesWriterMap map[string]*pterm.SpinnerPrinter,
	gph *apiv1.Graph,
) {
	resourceMap := make(map[string]apiv1.Resource)
	ioWriterMap := make(map[string]io.Writer)
	toBeWatched := apiv1.Resources{}

	// Get the resources to be watched.
	for _, res := range rel.Spec.Resources {
		if changes.ChangeOrder.ChangeSteps[res.ResourceKey()].Action != models.UnChanged {
			resourceMap[res.ResourceKey()] = res
			toBeWatched = append(toBeWatched, res)
		}
	}

	go func() {
		defer func() {
			if p := recover(); p != nil {
				cmdutil.RecoverErr(err)
				log.Error(*err)
			}
			if *err != nil {
				if !releaseCreated {
					return
				}
				release.UpdateReleasePhase(rel, apiv1.ReleasePhaseFailed, relLock)
				_ = release.UpdateApplyRelease(releaseStorage, rel, dryRun, relLock)
			}

			watchErrCh <- *err
		}()
		// Init the runtimes according to the resource types.
		runtimes, s := runtimeinit.Runtimes(*rel.Spec, *rel.State)
		if v1.IsErr(s) {
			panic(fmt.Errorf("failed to init runtimes: %s", s.String()))
		}

		// Prepare the tables for printing the details of the resources.
		tables := make(map[string]*printers.Table, len(toBeWatched))
		ticker := time.NewTicker(time.Millisecond * 100)
		defer ticker.Stop()

		// Record the watched and finished resources.
		watchedIDs := []string{}
		finished := make(map[string]bool)

		for !(len(finished) == len(toBeWatched)) {
			select {
			// Get the resource ID to be watched.
			case id := <-ac.WatchCh:
				res := resourceMap[id]
				// Set the timeout duration for watch context, here we set an experiential value of 60 minutes.
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*time.Duration(60))
				defer cancel()

				// Get the event channel for watching the resource.
				rsp := runtimes[res.Type].Watch(ctx, &runtime.WatchRequest{Resource: &res})
				if rsp == nil {
					log.Debug("unsupported resource type: %s", res.Type)
					continue
				}
				if v1.IsErr(rsp.Status) {
					panic(fmt.Errorf("failed to watch %s as %s", id, rsp.Status.String()))
				}

				w := rsp.Watchers
				table := printers.NewTable(w.IDs)
				tables[id] = table

				// Setup a go-routine to concurrently watch K8s and TF resources.
				if res.Type == apiv1.Kubernetes {
					healthPolicy, kind := getResourceInfo(&res)
					go watchK8sResources(id, kind, w.Watchers, table, tables, gph, dryRun, healthPolicy)
				} else if res.Type == apiv1.Terraform {
					go watchTFResources(id, w.TFWatcher, table, dryRun)
				} else {
					log.Debug("unsupported resource type to watch: %s", string(res.Type))
					continue
				}

				// Record the io writer related to the resource ID.
				ioWriterMap[id] = multi.NewWriter()
				watchedIDs = append(watchedIDs, id)

			// Refresh the tables printing details of the resources to be watched.
			default:
				for _, id := range watchedIDs {
					w, ok := ioWriterMap[id]
					if !ok {
						panic(fmt.Errorf("failed to get io writer while watching %s", id))
					}
					printTable(&w, id, tables)
				}
				for id, table := range tables {
					if finished[id] {
						continue
					}

					if table.AllCompleted() {
						finished[id] = true
						changesWriterMap[id].Success(fmt.Sprintf("Succeeded %s", pterm.Bold.Sprint(id)))

						// Update resource status to reconciled.
						resource := graph.FindGraphResourceByID(gph.Resources, id)
						if resource != nil {
							resource.Status = apiv1.Reconciled
						}
					}
				}
				<-ticker.C
			}
		}
	}()
}

// PortForward function will forward the specified port from local to the project Kubernetes Service.
//
// Example:
//
// o := newApplyOptions()
// spec, err := generate.GenerateSpecWithSpinner(o.RefProject, o.RefStack, o.RefWorkspace, nil, o.NoStyle)
//
//	if err != nil {
//		 return err
//	}
//
// err = PortForward(o, spec)
//
//	if err != nil {
//	  return err
//	}
//
// Fixme: gracefully exit when interrupted by SIGINT or SIGTERM.
func PortForward(
	o *ApplyOptions,
	spec *apiv1.Spec,
) error {
	if o.DryRun {
		fmt.Println("NOTE: Portforward doesn't work in DryRun mode")
		return nil
	}

	// portforward operation
	wo := &operation.PortForwardOperation{}
	if err := wo.PortForward(&operation.PortForwardRequest{
		Spec: spec,
		Port: o.PortForward,
	}); err != nil {
		return err
	}

	fmt.Println("Portforward has been completed!")
	return nil
}

type lineSummary struct {
	created, updated, deleted int
}

func (ls *lineSummary) Count(op models.ActionType) {
	switch op {
	case models.Create:
		ls.created++
	case models.Update:
		ls.updated++
	case models.Delete:
		ls.deleted++
	}
}

func allUnChange(changes *models.Changes) bool {
	for _, v := range changes.ChangeSteps {
		if v.Action != models.UnChanged {
			return false
		}
	}

	return true
}

func prompt(ui *terminal.UI) (string, error) {
	// don`t display yes item when only preview
	options := []string{"yes", "details", "no"}
	input, err := ui.InteractiveSelectPrinter.
		WithFilter(false).
		WithDefaultText(`Do you want to apply these diffs?`).
		WithOptions(options).
		WithDefaultOption("details").
		// To gracefully exit if interrupted by SIGINT or SIGTERM.
		WithOnInterruptFunc(func() {
			release.UpdateReleasePhase(rel, apiv1.ReleasePhaseFailed, relLock)
			release.UpdateApplyRelease(releaseStorage, rel, false, relLock)
			os.Exit(1)
		}).
		Show()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		return "", err
	}

	return input, nil
}

func watchK8sResources(
	id, kind string,
	chs []<-chan watch.Event,
	table *printers.Table,
	tables map[string]*printers.Table,
	gph *apiv1.Graph,
	dryRun bool,
	healthPolicy interface{},
) {
	defer func() {
		var err error
		if p := recover(); p != nil {
			cmdutil.RecoverErr(&err)
			log.Error(err)
		}
		if err != nil {
			if !releaseCreated {
				return
			}
			release.UpdateReleasePhase(rel, apiv1.ReleasePhaseFailed, relLock)
			_ = release.UpdateApplyRelease(releaseStorage, rel, dryRun, relLock)
		}
	}()

	// Set resource status to `reconcile failed` before reconcile successfully.
	resource := graph.FindGraphResourceByID(gph.Resources, id)
	if resource != nil {
		resource.Status = apiv1.ReconcileFail
	}

	// Resources selects
	cases := createSelectCases(chs)
	// Default select
	cases = append(cases, reflect.SelectCase{
		Dir:  reflect.SelectDefault,
		Chan: reflect.Value{},
		Send: reflect.Value{},
	})

	for {
		chosen, recv, recvOK := reflect.Select(cases)
		if cases[chosen].Dir == reflect.SelectDefault {
			continue
		}
		if recvOK {
			e := recv.Interface().(watch.Event)
			o := e.Object.(*unstructured.Unstructured)
			var detail string
			var ready bool
			if e.Type == watch.Deleted {
				detail = fmt.Sprintf("%s has beed deleted", o.GetName())
				ready = true
			} else {
				// Restore to actual type
				target := printers.Convert(o)
				// Check reconcile status with customized health policy for specific resource
				if healthPolicy != nil && kind == o.GetObjectKind().GroupVersionKind().Kind {
					if code, ok := kcl.ConvertKCLCode(healthPolicy); ok {
						resByte, err := yaml.Marshal(o.Object)
						if err != nil {
							log.Error(err)
							return
						}
						detail, ready = printers.PrintCustomizedHealthCheck(code, resByte)
					} else {
						detail, ready = printers.Generate(target)
					}
				} else {
					// Check reconcile status with default setup
					detail, ready = printers.Generate(target)
				}
			}

			// Mark ready for breaking loop
			if ready {
				e.Type = printers.READY
			}

			// Save watched msg
			table.Update(
				engine.BuildIDForKubernetes(o),
				printers.NewRow(e.Type, o.GetKind(), o.GetName(), detail))

			// Write back
			tables[id] = table
		}

		// Break when completed
		if table.AllCompleted() {
			break
		}
	}
}

func watchTFResources(
	id string,
	ch <-chan runtime.TFEvent,
	table *printers.Table,
	dryRun bool,
) {
	defer func() {
		var err error
		if p := recover(); p != nil {
			cmdutil.RecoverErr(&err)
			log.Error(err)
		}
		if err != nil {
			if !releaseCreated {
				return
			}
			release.UpdateReleasePhase(rel, apiv1.ReleasePhaseFailed, relLock)
			_ = release.UpdateApplyRelease(releaseStorage, rel, dryRun, relLock)
		}
	}()

	for {
		parts := strings.Split(id, engine.Separator)
		// A valid Terraform resource ID should consist of 4 parts, including the information of the provider type
		// and resource name, for example: hashicorp:random:random_password:example-dev-kawesome.
		if len(parts) != 4 {
			panic(fmt.Errorf("invalid Terraform resource id: %s", id))
		}

		tfEvent := <-ch
		if tfEvent == runtime.TFApplying {
			table.Update(
				id,
				printers.NewRow(watch.EventType("Applying"),
					strings.Join([]string{parts[1], parts[2]}, engine.Separator), parts[3], "Applying..."))
		} else if tfEvent == runtime.TFSucceeded {
			table.Update(
				id,
				printers.NewRow(printers.READY,
					strings.Join([]string{parts[1], parts[2]}, engine.Separator), parts[3], "Apply succeeded"))
		} else {
			table.Update(
				id,
				printers.NewRow(watch.EventType("Failed"),
					strings.Join([]string{parts[1], parts[2]}, engine.Separator), parts[3], "Apply failed"))
		}

		// Break when all completed.
		if table.AllCompleted() {
			break
		}
	}
}

func createSelectCases(chs []<-chan watch.Event) []reflect.SelectCase {
	cases := make([]reflect.SelectCase, 0, len(chs))
	for _, ch := range chs {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		})
	}
	return cases
}

func printTable(w *io.Writer, id string, tables map[string]*printers.Table) {
	// Reset the buffer for live flushing.
	(*w).(*bytes.Buffer).Reset()

	// Print resource Key as heading text
	_, _ = fmt.Fprintln(*w, pretty.LightCyanBold("[%s]", id))

	table, ok := tables[id]
	if !ok {
		// Unsupported resource, leave a hint
		_, _ = fmt.Fprintln(*w, "Skip monitoring unsupported resources")
	} else {
		// Print table
		data := table.Print()
		_ = pterm.DefaultTable.
			WithStyle(pterm.NewStyle(pterm.FgDefault)).
			WithHeaderStyle(pterm.NewStyle(pterm.FgDefault)).
			WithHasHeader().WithSeparator("  ").WithData(data).WithWriter(*w).Render()
	}
}

// getResourceInfo get health policy and kind from resource for customized health check purpose
func getResourceInfo(res *apiv1.Resource) (healthPolicy interface{}, kind string) {
	var ok bool
	if res.Extensions != nil {
		healthPolicy = res.Extensions[apiv1.FieldHealthPolicy]
	}
	if res.Attributes == nil {
		panic(fmt.Errorf("resource has no Attributes field in the Spec: %s", res))
	}
	if kind, ok = res.Attributes[apiv1.FieldKind].(string); !ok {
		panic(fmt.Errorf("failed to get kind from resource attributes: %s", res.Attributes))
	}
	return healthPolicy, kind
}

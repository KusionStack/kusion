package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/liu-hm19/pterm"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/printers"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/engine/resource/graph"
	"kusionstack.io/kusion/pkg/engine/runtime"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/infra/util/semaphore"
	"kusionstack.io/kusion/pkg/log"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
	"kusionstack.io/kusion/pkg/util/kcl"
	"kusionstack.io/kusion/pkg/util/pretty"
)

// To handle the release phase update when panic occurs.
// Fixme: adopt a more centralized approach to manage the release update before exiting, instead of
// scattering them across different go-routines.
var (
	relLock        = &sync.Mutex{}
	releaseCreated = false
	releaseStorage release.Storage
)

// The Apply function will apply the resources changes
// through the execution Kusion Engine, and will save
// the state to specified storage.
//
// You can customize the runtime of engine and the state
// storage through `runtime` and `storage` parameters.
func Apply(
	ctx context.Context,
	o *APIOptions,
	storage release.Storage,
	rel *apiv1.Release,
	gph *apiv1.Graph,
	changes *models.Changes,
	out io.Writer,
) (*apiv1.Release, error) {
	logger := logutil.GetLogger(ctx)
	var err error

	// construct the apply operation
	ac := &operation.ApplyOperation{
		Operation: models.Operation{
			Stack:          changes.Stack(),
			ReleaseStorage: storage,
			MsgCh:          make(chan models.Message),
			IgnoreFields:   o.IgnoreFields,
			Sem:            semaphore.New(int64(o.MaxConcurrent)),
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

	// progress bar, print dag walk detail
	progressbar, err := pterm.DefaultProgressbar.
		WithMaxWidth(0). // Set to 0, the terminal width will be used
		WithTotal(len(changes.StepKeys)).
		WithWriter(out).
		WithRemoveWhenDone().
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
		gph,
		rel,
	)

	watchErrCh := make(chan error)
	// Apply while watching the resources.
	if o.Watch && !o.DryRun {
		logger.Info("Start watching resources ...")
		Watch(
			ac,
			changes,
			&err,
			o.DryRun,
			watchErrCh,
			multi,
			changesWriterMap,
			gph,
			rel,
		)
		logger.Info("Watch completed ...")
	}

	var upRel *apiv1.Release
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
			return nil, fmt.Errorf("apply failed, status:\n%v", st)
		}
		upRel = rsp.Release
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
	logger.Info(fmt.Sprintf("Apply complete! Resources: %d created, %d updated, %d deleted.", ls.created, ls.updated, ls.deleted))
	return upRel, nil
}

// Watch function will observe the changes of each resource
// by the execution engine.
//
// Example:
//
//	o := NewApplyOptions()
//	kubernetesRuntime, err := runtime.NewKubernetesRuntime()
//	if err != nil {
//	    return err
//	}
//
//	Watch(o, kubernetesRuntime, planResources, changes, os.Stdout)
//	if err != nil {
//	    return err
//	}
//
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
	rel *apiv1.Release,
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
			log.Debug("entering defer func() for watch()")
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
				log.Debug("entering case id := <-ac.WatchCh")
				log.Debug("id", id)
				res := resourceMap[id]
				log.Debug("res.Type", res.Type)
				// Set the timeout duration for watch context, here we set an experiential value of 60 minutes.
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(120))
				defer cancel()

				// Get the event channel for watching the resource.
				rsp := runtimes[res.Type].Watch(ctx, &runtime.WatchRequest{Resource: &res})
				log.Debug("rsp", rsp)
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
					healthPolicy, kind := getHealthPolicy(&res)
					log.Debug("healthPolicyhealthPolicyhealthPolicyhealthPolicyhealthPolicy", healthPolicy)
					go watchK8sResources(id, kind, w.Watchers, table, tables, gph, dryRun, healthPolicy, rel)
				} else if res.Type == apiv1.Terraform {
					go watchTFResources(id, w.TFWatcher, table, dryRun, rel)
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
	gph *apiv1.Graph,
	rel *apiv1.Release,
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

// getHealthPolicy get health policy and kind from resource for customized health check purpose
func getHealthPolicy(res *apiv1.Resource) (healthPolicy interface{}, kind string) {
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

func watchK8sResources(
	id, kind string,
	chs []<-chan watch.Event,
	table *printers.Table,
	tables map[string]*printers.Table,
	gph *apiv1.Graph,
	dryRun bool,
	healthPolicy interface{},
	rel *apiv1.Release,
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
					log.Debug("healthPolicy", healthPolicy)
					if code, ok := kcl.ConvertKCLCode(healthPolicy); ok {
						resByte, err := yaml.Marshal(o.Object)
						log.Debug("kcl health policy code", code)
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
	rel *apiv1.Release,
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

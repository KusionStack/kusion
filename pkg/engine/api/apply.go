package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
	"time"

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
	"kusionstack.io/kusion/pkg/server/middleware"
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
	sysLogger := logutil.GetLogger(ctx)
	runLogger := logutil.GetRunLogger(ctx)
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

	// Prepare the writer to print the operation progress and results.
	for _, key := range changes.Values() {
		logutil.LogToAll(sysLogger, runLogger, "Info", fmt.Sprintf("Pending %s", key.ID))
	}

	// wait msgCh close
	var wg sync.WaitGroup
	// receive msg and print detail
	go ProcessApplyDetails(
		ctx,
		ac,
		&err,
		&wg,
		changes,
		&ls,
		o.DryRun,
		gph.Resources,
		gph,
		rel,
	)

	watchErrCh := make(chan error)
	// Apply while watching the resources.
	if o.Watch && !o.DryRun {
		logutil.LogToAll(sysLogger, runLogger, "Info", fmt.Sprintf("Start watching resources with timeout %d seconds ...", o.WatchTimeout))
		Watch(
			ctx,
			ac,
			changes,
			&err,
			o.DryRun,
			watchErrCh,
			o.WatchTimeout,
			gph,
			rel,
		)
		logutil.LogToAll(sysLogger, runLogger, "Info", "Watch started ...")
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
	logutil.LogToAll(sysLogger, runLogger, "Info", fmt.Sprintf("Apply complete! Resources: %d created, %d updated, %d deleted.", ls.created, ls.updated, ls.deleted))
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
	ctx context.Context,
	ac *operation.ApplyOperation,
	changes *models.Changes,
	err *error,
	dryRun bool,
	watchErrCh chan error,
	watchTimeout int,
	gph *apiv1.Graph,
	rel *apiv1.Release,
) {
	resourceMap := make(map[string]apiv1.Resource)
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
		// Get syslogger
		sysLogger := logutil.GetLogger(ctx)
		runLogger := logutil.GetRunLogger(ctx)
		// Init the runtimes according to the resource types.
		runtimes, s := runtimeinit.Runtimes(*rel.Spec, *rel.State)
		if v1.IsErr(s) {
			panic(fmt.Errorf("failed to init runtimes: %s", s.String()))
		}

		// Prepare the tables for printing the details of the resources.
		ticker := time.NewTicker(time.Millisecond * 1000)
		defer ticker.Stop()

		// Record the watched and finished resources.
		watching := make(map[string]bool)
		finished := make(map[string]bool)

		logutil.LogToAll(sysLogger, runLogger, "Info", "Total resources to watch: ", len(toBeWatched))
		for !(len(finished) == len(toBeWatched)) {
			select {
			// Get the resource ID to be watched.
			case id := <-ac.WatchCh:
				res := resourceMap[id]
				// Set the timeout duration for watch context, here we set an experiential value of 60 minutes.
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(watchTimeout))
				ctx = context.WithValue(ctx, middleware.APILoggerKey, sysLogger)
				ctx = context.WithValue(ctx, middleware.RunLoggerKey, runLogger)
				defer cancel()

				// Get the event channel for watching the resource.
				rsp := runtimes[res.Type].Watch(ctx, &runtime.WatchRequest{Resource: &res})
				logutil.LogToAll(sysLogger, runLogger, "Info", fmt.Sprintf("Watching resource rsp: %v", rsp))
				if rsp == nil {
					log.Debug("unsupported resource type: %s", res.Type)
					continue
				}
				if v1.IsErr(rsp.Status) {
					panic(fmt.Errorf("failed to watch %s as %s", id, rsp.Status.String()))
				}

				w := rsp.Watchers
				logutil.LogToAll(sysLogger, runLogger, "Info", "setting finished to false...", "id", id, "timeElapsed", time.Now().String())
				watching[id] = false

				// Setup a go-routine to concurrently watch K8s and TF resources.
				if res.Type == apiv1.Kubernetes {
					healthPolicy, kind := getHealthPolicy(&res)
					go watchK8sResources(ctx, id, kind, w.Watchers, watching, gph, dryRun, healthPolicy, rel)
				} else if res.Type == apiv1.Terraform {
					go watchTFResources(ctx, id, w.TFWatcher, watching, dryRun, rel)
				} else {
					log.Debug("unsupported resource type to watch: %s", string(res.Type))
					continue
				}
			// Refresh the tables printing details of the resources to be watched.
			default:
				for id := range watching {
					logutil.LogToAll(sysLogger, runLogger, "Info", "watching resource...", "id", id, "timeElapsed", time.Now().String())
					if watching[id] {
						finished[id] = true
						logutil.LogToAll(sysLogger, runLogger, "Info", "finished watching. resource is reconciled...", "id", id, "timeElapsed", time.Now().String())
						delete(watching, id)
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

// ProcessApplyDetails function will receive the messages of the apply operation and process the details.
// Fixme: abstract the input variables into a struct.
func ProcessApplyDetails(
	ctx context.Context,
	ac *operation.ApplyOperation,
	err *error,
	wg *sync.WaitGroup,
	changes *models.Changes,
	ls *lineSummary,
	dryRun bool,
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
	}()
	sysLogger := logutil.GetLogger(ctx)
	runLogger := logutil.GetRunLogger(ctx)
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
					title = fmt.Sprintf("Skipped %s", changeStep.ID)
					logutil.LogToAll(sysLogger, runLogger, "Info", title)
				} else {
					logutil.LogToAll(sysLogger, runLogger, "Info", fmt.Sprintf("Succeeded %s", msg.ResourceID))
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

				ls.Count(changeStep.Action)
			case models.Failed:
				errStr := pretty.ErrorT.Sprintf("apply %s failed as: %s\n", msg.ResourceID, msg.OpErr.Error())
				logutil.LogToAll(sysLogger, runLogger, "Error", errStr)
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
					changeStep.ID,
				)
				logutil.LogToAll(sysLogger, runLogger, "Info", title)
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
	ctx context.Context,
	id, kind string,
	chs []<-chan watch.Event,
	watching map[string]bool,
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

	sysLogger := logutil.GetLogger(ctx)
	runLogger := logutil.GetRunLogger(ctx)

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

	// when both are ready, break the loop
	ready := map[string]bool{}

	for {
		chosen, recv, recvOK := reflect.Select(cases)
		if cases[chosen].Dir == reflect.SelectDefault {
			// handle timeout
			select {
			case <-ctx.Done():
				logutil.LogToAll(sysLogger, runLogger, "Info", "Watch timeout reached. Setting watching to true...", "id", id, "timeElapsed", time.Now().String())
				watching[id] = true
				return
			default:
				continue
			}
		}
		if recvOK {
			e := recv.Interface().(watch.Event)
			o := e.Object.(*unstructured.Unstructured)
			// var detail string
			if e.Type == watch.Deleted {
				ready["custom"] = true
			} else {
				// Restore to actual type
				target := printers.Convert(o)
				// Check reconcile status with customized health policy for specific resource
				if healthPolicy != nil && kind == o.GetObjectKind().GroupVersionKind().Kind {
					ready["custom"] = false
					if code, ok := kcl.ConvertKCLCode(healthPolicy); ok {
						resByte, err := yaml.Marshal(o.Object)
						if err != nil {
							log.Error(err)
							return
						}
						kclResp, kclReady := printers.PrintCustomizedHealthCheck(code, resByte)
						if kclReady {
							ready["custom"] = true
							logutil.LogToAll(sysLogger, runLogger, "Info", "Customized health check ready: ", "kclResp", kclResp, "timeElapsed", time.Now().String(), "id", id)
						}
					} else {
						// Check reconcile status with default setup
						ready["default"] = false
						_, defaultReady := printers.Generate(target)
						if defaultReady {
							ready["default"] = true
							logutil.LogToAll(sysLogger, runLogger, "Info", "Customized health check had a problem. Default health check ready: ", "timeElapsed", time.Now().String(), "id", id)
						}
						continue
					}
				} else {
					// Check reconcile status with default setup
					ready["default"] = false
					_, defaultReady := printers.Generate(target)
					if defaultReady {
						ready["default"] = true
						logutil.LogToAll(sysLogger, runLogger, "Info", "default health check ready: ", "timeElapsed", time.Now().String(), "id", id)
					}
					continue
				}
			}
		}
		// Mark ready for breaking loop
		if allReady(ready) {
			logutil.LogToAll(sysLogger, runLogger, "Info", "Kubernetes resource reconciled. Setting finished to true...", "id", id, "timeElapsed", time.Now().String())
			watching[id] = true
			break
		}
	}
}

func watchTFResources(
	ctx context.Context,
	id string,
	ch <-chan runtime.TFEvent,
	watching map[string]bool,
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
	sysLogger := logutil.GetLogger(ctx)
	runLogger := logutil.GetRunLogger(ctx)

	var ready bool
	for {
		parts := strings.Split(id, engine.Separator)
		// A valid Terraform resource ID should consist of 4 parts, including the information of the provider type
		// and resource name, for example: hashicorp:random:random_password:example-dev-kawesome.
		if len(parts) != 4 {
			panic(fmt.Errorf("invalid Terraform resource id: %s", id))
		}

		tfEvent := <-ch
		if tfEvent == runtime.TFApplying {
			continue
		} else {
			ready = true
		}

		// Mark ready for breaking loop
		if ready {
			watching[id] = true
			logutil.LogToAll(sysLogger, runLogger, "Info", "Terraform resource apply completed. Setting finished to true...", "id", id, "timeElapsed", time.Now().String())
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

func allReady(ready map[string]bool) bool {
	for _, r := range ready {
		if !r {
			return false
		}
	}
	return true
}

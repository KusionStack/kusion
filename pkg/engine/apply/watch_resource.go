package apply

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/liu-hm19/pterm"
	"gopkg.in/yaml.v3"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/engine"
	applystate "kusionstack.io/kusion/pkg/engine/apply/state"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/printers"
	"kusionstack.io/kusion/pkg/engine/resource/graph"
	"kusionstack.io/kusion/pkg/engine/runtime"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util/kcl"
	"kusionstack.io/kusion/pkg/util/pretty"
)

// Watch function will watch the changed Kubernetes and Terraform resources.
func Watch(
	state *applystate.State,
	watchResult chan error,
	watchChan chan string,
	multi *pterm.MultiPrinter,
	changesWriterMap map[string]*pterm.SpinnerPrinter,
	changes *models.Changes,
) {
	var err error
	defer func() {
		watchResult <- err
		close(watchResult)
	}()
	defer cmdutil.RecoverErr(&err)

	resourceMap := make(map[string]apiv1.Resource)
	ioWriterMap := make(map[string]io.Writer)
	toBeWatched := apiv1.Resources{}

	// Get the resources to be watched.
	for _, res := range state.TargetRel.Spec.Resources {
		if changes.ChangeOrder.ChangeSteps[res.ResourceKey()].Action != models.UnChanged {
			resourceMap[res.ResourceKey()] = res
			toBeWatched = append(toBeWatched, res)
		}
	}

	// Init the runtimes according to the resource types.
	runtimes, s := runtimeinit.Runtimes(*state.TargetRel.Spec, *state.TargetRel.State)
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

	// Resource error chan
	watchErrChan := make(chan error)
	defer close(watchErrChan)

	for !(len(finished) == len(toBeWatched)) {
		select {
		// Get the resource ID to be watched.
		case id := <-watchChan:
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
				go watchK8sResources(id, kind, w.Watchers, table, tables, healthPolicy, state.Gph.Resources, watchErrChan)
			} else if res.Type == apiv1.Terraform {
				go watchTFResources(id, w.TFWatcher, table, watchErrChan)
			} else {
				log.Debug("unsupported resource type to watch: %s", string(res.Type))
				continue
			}

			// Record the io writer related to the resource ID.
			ioWriterMap[id] = multi.NewWriter()
			watchedIDs = append(watchedIDs, id)

		case err = <-watchErrChan:
			if err != nil {
				return
			}

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
					resource := graph.FindGraphResourceByID(state.Gph.Resources, id)
					if resource != nil {
						resource.Status = apiv1.Reconciled
					}
				}
			}
			<-ticker.C
		}
	}
}

func watchTFResources(
	id string,
	ch <-chan runtime.TFEvent,
	table *printers.Table,
	errChan chan<- error,
) {
	var err error
	defer func() {
		if err != nil {
			errChan <- WrappedErr(err, "watchTFResources err")
			close(errChan)
		}
	}()

	defer cmdutil.RecoverErr(&err)

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

func watchK8sResources(
	id, kind string,
	chs []<-chan watch.Event,
	table *printers.Table,
	tables map[string]*printers.Table,
	healthPolicy interface{},
	gphResource *apiv1.GraphResources,
	errChan chan<- error,
) {
	var err error
	defer func() {
		if err != nil {
			errChan <- WrappedErr(err, "watchK8sResources err")
			close(errChan)
		}
	}()

	defer cmdutil.RecoverErr(&err)

	// Set resource status to `reconcile failed` before reconcile successfully.
	resource := graph.FindGraphResourceByID(gphResource, id)
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
						var resByte []byte
						resByte, err = yaml.Marshal(o.Object)
						if err != nil {
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

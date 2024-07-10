package operation

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/howieyuen/uilive"
	"github.com/liu-hm19/pterm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/printers"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/engine/runtime"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util/pretty"
)

type WatchOperation struct {
	models.Operation
}

type WatchRequest struct {
	models.Request
	Spec *apiv1.Spec
}

func (wo *WatchOperation) Watch(req *WatchRequest) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := validateWatchRequest(req); err != nil {
		return err
	}

	// init runtimes
	resources := req.Spec.Resources
	runtimes, s := runtimeinit.Runtimes(*req.Spec)
	if v1.IsErr(s) {
		return errors.New(s.Message())
	}
	wo.RuntimeMap = runtimes

	// Result channels
	msgChs := make(map[string]*runtime.SequentialWatchers, len(resources))
	// Keep sorted
	ids := make([]string, resources.Len())
	// Collect watchers
	for i := range resources {
		res := &resources[i]
		t := res.Type

		// Save id first, might have TF resources
		ids[i] = res.ResourceKey()

		// Get watchers, only support k8s resources
		resp := runtimes[t].Watch(ctx, &runtime.WatchRequest{Resource: res})
		if resp == nil {
			log.Debug("unsupported resource type: %s", t)
			continue
		}
		if v1.IsErr(resp.Status) {
			return fmt.Errorf(resp.Status.String())
		}

		// Save watchers
		msgChs[res.ResourceKey()] = resp.Watchers
	}

	// Console writer
	writer := uilive.New()
	writer.RefreshInterval = time.Second * 10
	writer.Start()
	defer writer.Stop()

	// Table data
	tables := make(map[string]*printers.Table, len(ids))
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// Counting completed resource
	finished := make(map[string]bool)

	// Start go routine for each table
	for _, id := range ids {
		sw, ok := msgChs[id]
		if !ok { // Terraform resource, skip
			continue
		}
		// New target table
		table := printers.NewTable(sw.IDs)
		// Save tables first
		tables[id] = table
		// Start watching resource
		go func(id string, chs []<-chan watch.Event, table *printers.Table) {
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
						detail, ready = printers.Generate(target)
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
		}(id, sw.Watchers, table)
	}

	// No k8s resources
	if len(tables) == 0 {
		wo.printTables(writer, ids, tables)
		return nil
	}

	// Waiting for all tables completed
	for {
		// Finish watch
		if len(finished) == len(tables) {
			break
		}

		// Range tables
		for id, table := range tables {
			// All channels are isCompleted
			if table.AllCompleted() {
				finished[id] = true
			}
		}

		// Render table every 1s
		<-ticker.C
		wo.printTables(writer, ids, tables)
	}
	return nil
}

func (wo *WatchOperation) printTables(w *uilive.Writer, ids []string, tables map[string]*printers.Table) {
	for i, id := range ids {
		// Print resource Key as heading text
		_, _ = fmt.Fprintln(w, pretty.LightCyanBold("[%s]", id))

		table, ok := tables[id]
		if !ok {
			// Terraform resource, leave a hint
			_, _ = fmt.Fprintln(w, "Skip monitoring Terraform resources")
		} else {
			// Print table
			data := table.Print()
			_ = pterm.DefaultTable.WithHasHeader().WithSeparator("  ").WithData(data).WithWriter(w).Render()
		}

		// Split each resource with blank line
		if i != len(ids)-1 {
			_, _ = fmt.Fprintln(w)
		}
	}

	_ = w.Flush()
}

func validateWatchRequest(req *WatchRequest) error {
	if req == nil {
		return errors.New("request is nil")
	}
	if err := release.ValidateSpec(req.Spec); err != nil {
		return err
	}
	return nil
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

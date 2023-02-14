package operation

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/gosuri/uilive"
	"github.com/pterm/pterm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8swatch "k8s.io/apimachinery/pkg/watch"

	"kusionstack.io/kusion/pkg/engine"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/printers"
	"kusionstack.io/kusion/pkg/engine/runtime"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/status"
	"kusionstack.io/kusion/pkg/util/pretty"
)

type WatchOperation struct {
	opsmodels.Operation
}

type WatchRequest struct {
	opsmodels.Request `json:",inline" yaml:",inline"`
}

func (wo *WatchOperation) Watch(req *WatchRequest) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// init runtimes
	resources := req.Spec.Resources
	runtimes, s := runtimeinit.Runtimes(resources)
	if status.IsErr(s) {
		return errors.New(s.Message())
	}
	wo.RuntimeMap = runtimes

	// Result channels
	msgChs := make(map[string][]<-chan k8swatch.Event, len(resources))
	// Keep sorted
	ids := make([]string, resources.Len())
	// Collect watchers
	for i := range resources {
		res := &resources[i]

		// only support k8s resources
		t := res.Type
		if runtime.Kubernetes != t {
			return fmt.Errorf("WARNING: Watch only support Kubernetes resources for now")
		}

		// Get watchers
		resp := runtimes[t].Watch(ctx, &runtime.WatchRequest{Resource: res})
		if status.IsErr(resp.Status) {
			return fmt.Errorf(resp.Status.String())
		}
		// Save id
		ids[i] = res.ResourceKey()
		// Save channels
		msgChs[res.ResourceKey()] = resp.ResultChs
	}

	// Console writer
	writer := uilive.New()
	writer.RefreshInterval = time.Minute * 1
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
		chs, ok := msgChs[id]
		if !ok {
			continue
		}
		// Get or new the target table
		table, exist := tables[id]
		if !exist {
			table = printers.NewTable(len(chs))
		}
		go func(id string, chs []<-chan k8swatch.Event, table *printers.Table) {
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
					e := recv.Interface().(k8swatch.Event)
					o := e.Object.(*unstructured.Unstructured)
					var detail string
					var ready bool
					if e.Type == k8swatch.Deleted {
						detail = fmt.Sprintf("%s has beed deleted", o.GetName())
						ready = true
					} else {
						// Restore to actual type
						target := printers.Convert(o)
						detail, ready = printers.TG.GenerateTable(target)
					}

					// Mark ready for breaking loop
					if ready {
						e.Type = printers.READY
					}

					// Save watched msg
					table.InsertOrUpdate(
						engine.BuildIDForKubernetes(o.GetAPIVersion(), o.GetKind(), o.GetNamespace(), o.GetName()),
						printers.NewRow(e.Type, o.GetKind(), o.GetName(), detail))

					// Write back
					tables[id] = table
				}

				// Break when completed
				if table.IsCompleted() {
					break
				}
			}
		}(id, chs, table)
	}

	// Waiting for all tables completed
	for {
		// Finish watch
		if len(finished) == len(ids) {
			break
		}

		// Range tables
		for id, table := range tables {
			// All channels are isCompleted
			if table.IsCompleted() {
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
		_, _ = fmt.Fprintf(w, "%s\n", pretty.LightCyanBold("[%s]", id))

		table, ok := tables[id]
		if !ok {
			continue
		}
		// Print table
		data := table.Print()
		_ = pterm.DefaultTable.WithHasHeader().WithSeparator("  ").WithData(data).WithWriter(w).Render()

		// Split each resource with blank line
		if i != len(ids)-1 {
			_, _ = fmt.Fprintln(w)
		}
	}

	_ = w.Flush()
}

func createSelectCases(chs []<-chan k8swatch.Event) []reflect.SelectCase {
	cases := make([]reflect.SelectCase, 0, len(chs))
	for _, ch := range chs {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		})
	}
	return cases
}

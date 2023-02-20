package operation

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/howieyuen/uilive"
	"github.com/pterm/pterm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8swatch "k8s.io/apimachinery/pkg/watch"

	"kusionstack.io/kusion/pkg/engine"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/printers"
	"kusionstack.io/kusion/pkg/engine/runtime"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/log"
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
	msgChs := make(map[string]*runtime.SequentialWatchers, len(resources))
	// Keep sorted
	ids := make([]string, resources.Len())
	// Collect watchers
	for i := range resources {
		res := &resources[i]
		t := res.Type

		// Get watchers, only support k8s resources
		resp := runtimes[t].Watch(ctx, &runtime.WatchRequest{Resource: res})
		if resp == nil {
			log.Debug("unsupported resource type: %s", t)
			continue
		}
		if status.IsErr(resp.Status) {
			return fmt.Errorf(resp.Status.String())
		}

		// Save id
		ids[i] = res.ResourceKey()
		// Save channels
		msgChs[res.ResourceKey()] = resp.Watchers
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
		sw, ok := msgChs[id]
		if !ok {
			continue
		}
		// Get or new the target table
		table, exist := tables[id]
		if !exist {
			table = printers.NewTable(sw.IDs)
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

	// Waiting for all tables completed
	for {
		// Finish watch
		if len(finished) == len(ids) {
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

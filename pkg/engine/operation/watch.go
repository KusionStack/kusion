package operation

import (
	"context"
	"fmt"
	"time"

	"github.com/gosuri/uilive"
	"github.com/pterm/pterm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8swatch "k8s.io/apimachinery/pkg/watch"

	"kusionstack.io/kusion/pkg/engine"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/printers"
	"kusionstack.io/kusion/pkg/engine/printers/k8s"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/status"
	"kusionstack.io/kusion/pkg/util/pretty"
)

var tableGenerator = printers.NewTableGenerator().With(k8s.AddHandlers)

type WatchOperation struct {
	Runtime runtime.Runtime
}

type WatchRequest struct {
	opsmodels.Request `json:",inline" yaml:",inline"`
}

func (wo *WatchOperation) Watch(req *WatchRequest) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resources := req.Spec.Resources
	// Result channels
	msgChs := make(map[string][]<-chan k8swatch.Event, len(resources))
	// Keep sorted
	ids := make([]string, resources.Len())
	// Collect watchers
	for i := range resources {
		res := &resources[i]
		// Get watchers
		resp := wo.Runtime.Watch(ctx, &runtime.WatchRequest{Resource: res})
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
	finished := make(map[string]bool, len(ids))
	// Start printing
	for {
		// Finish watch
		if len(finished) == len(ids) {
			break
		}
		// Range tables by id
		for _, id := range ids {
			chs, ok := msgChs[id]
			if !ok {
				continue
			}

			// Get or new the target table
			table, exist := tables[id]
			if !exist {
				table = printers.NewTable()
			}

			// Range channels for each table
			for _, ch := range chs {
				select {
				case e := <-ch:
					o := e.Object.(*unstructured.Unstructured)
					var detail string
					var ready bool
					if e.Type == k8swatch.Deleted {
						detail = fmt.Sprintf("%s has beed deleted", o.GetName())
						ready = true
					} else {
						// Restore to actual type
						target := k8s.Convert(o)
						detail, ready = tableGenerator.GenerateTable(target)
					}

					// Mark ready for breaking loop
					if ready {
						e.Type = printers.READY
					}

					// Save watched msg
					table.InsertOrUpdate(
						engine.BuildIDForKubernetes(o.GetAPIVersion(), o.GetKind(), o.GetNamespace(), o.GetName()),
						printers.NewRow(e.Type, o.GetKind(), o.GetName(), detail))
				case <-ticker.C:
					// Should never reach
				}
			}

			// All channels are isCompleted
			if table.IsCompleted() {
				finished[id] = true
			}

			// Write back
			tables[id] = table
		}
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

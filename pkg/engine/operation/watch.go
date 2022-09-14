package operation

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gosuri/uilive"
	k8sWatch "k8s.io/apimachinery/pkg/watch"

	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/status"
	"kusionstack.io/kusion/pkg/util/pretty"
)

type WatchOperation struct {
	Runtime runtime.Runtime
}

type WatchRequest struct {
	opsmodels.Request `json:",inline" yaml:",inline"`
}

func (wo *WatchOperation) Watch(req *WatchRequest) error {
	ctx := context.Background()
	resources := req.Spec.Resources
	// Watched channels
	msgChs := make([]<-chan runtime.RowData, resources.Len())
	// Sorted ids
	ids := make([]string, resources.Len())
	for i := range resources {
		resp := wo.Runtime.Watch(ctx, &runtime.WatchRequest{Resource: &resources[i]})
		if status.IsErr(resp.Status) {
			return fmt.Errorf(resp.Status.String())
		}
		ids[i] = resources[i].ResourceKey()
		msgChs[i] = resp.MsgCh
	}

	writer := uilive.New()
	writer.RefreshInterval = time.Minute * 1
	writer.Start()
	defer writer.Stop()

	table := make(map[string][]runtime.RowData, resources.Len())
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		closed := 0
		for i, ch := range msgChs {
			key := ids[i]
			select {
			case msg, ok := <-ch:
				if !ok {
					closed++
					continue
				}
				// Save watching details
				table[key] = append(table[key], msg)
			case <-ticker.C:
			}
		}
		if closed == len(msgChs) {
			break
		}
		wo.printTable(writer, ids, table)
	}
	return nil
}

func (wo *WatchOperation) printTable(w *uilive.Writer, ids []string, table map[string][]runtime.RowData) {
	for i, id := range ids {
		if table[id] == nil {
			continue
		}
		// Print resource Key as heading text
		fmt.Fprintf(w, "[%s]\n", pretty.CyanBold(id))
		for _, row := range table[id] {
			printRow(w, row)
		}
		// Split each resource
		if i != len(ids)-1 {
			fmt.Fprintln(w)
		}
	}

	w.Flush()
}

func printRow(w io.Writer, item runtime.RowData) {
	eventType := item.Event.Type
	var eventTypeS string
	switch eventType {
	case k8sWatch.Added:
		eventTypeS = pretty.Green(string(eventType))
	case k8sWatch.Deleted:
		eventTypeS = pretty.Red(string(eventType))
	case k8sWatch.Modified:
		eventTypeS = pretty.Yellow(string(eventType))
	case k8sWatch.Error:
		eventTypeS = pretty.Red(string(eventType))
	default:
		eventTypeS = pretty.Cyan(string(eventType))
	}
	fmt.Fprintf(w, "    [%s] %s\n", eventTypeS, item.Message)
}

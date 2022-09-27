package operation

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gosuri/uilive"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8swatch "k8s.io/apimachinery/pkg/watch"

	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/models"
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

type rowData struct {
	Type    k8swatch.EventType
	Message string
}

func (wo *WatchOperation) Watch(req *WatchRequest) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resources := req.Spec.Resources
	// Result channels
	msgChs := make(map[string]<-chan k8swatch.Event)
	// Keep sorted
	ids := make([]string, 0, resources.Len())
	// Collect watchers and sub watchers
	collectWatcher := func(res *models.Resource) error {
		resp := wo.Runtime.Watch(ctx, &runtime.WatchRequest{Resource: res})
		if status.IsErr(resp.Status) {
			return fmt.Errorf(resp.Status.String())
		}
		ids = append(ids, res.ResourceKey())
		msgChs[res.ResourceKey()] = resp.ResultCh
		return nil
	}

	for i := range resources {
		if err := collectWatcher(&resources[i]); err != nil {
			return err
		}

		// Build sub watcher if had
		sub, has := hasDependents(&resources[i])
		if has {
			if err := collectWatcher(sub); err != nil {
				return err
			}
		}
	}

	writer := uilive.New()
	writer.RefreshInterval = time.Minute * 1
	writer.Start()
	defer writer.Stop()

	table := make(map[string][]rowData, len(msgChs))
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// Counting ready resource
	finished := 0
	// Start printing
	for {
		// Finish watch
		if finished == len(ids) {
			break
		}
		for _, key := range ids {
			ch := msgChs[key]
			if ch == nil {
				continue
			}
			select {
			case e := <-ch:
				o := e.Object.(*unstructured.Unstructured)
				var msg string
				var ready bool
				if e.Type == k8swatch.Deleted {
					msg = fmt.Sprintf("%s has beed deleted", o.GetName())
					ready = true
				} else {
					// Restore to actual type
					target := k8s.Convert(o)
					msg, ready = tableGenerator.GenerateTable(target)
				}
				if ready {
					msg += fmt.Sprintf("\n%s has been synced already.", key)
					// Remove watchers
					msgChs[key] = nil
					finished++
				}

				// Save watched data
				table[key] = append(table[key], rowData{Type: e.Type, Message: msg})
			case <-ticker.C:
				// Should never reach
			}
		}
		wo.printTable(writer, ids, table)
	}
	return nil
}

func (wo *WatchOperation) printTable(w *uilive.Writer, ids []string, table map[string][]rowData) {
	for i, id := range ids {
		// Print resource Key as heading text
		fmt.Fprintf(w, "[%s]\n", pretty.CyanBold(id))

		if table[id] == nil {
			continue
		}
		// Print each event as a row
		for _, row := range table[id] {
			printRow(w, row)
		}

		// Split each resource with blank line
		if i != len(ids)-1 {
			fmt.Fprintln(w)
		}
	}

	w.Flush()
}

func printRow(w io.Writer, row rowData) {
	eventType := row.Type
	var eventTypeS string
	switch eventType {
	case k8swatch.Added:
		eventTypeS = pretty.Green(string(eventType))
	case k8swatch.Deleted:
		eventTypeS = pretty.Red(string(eventType))
	case k8swatch.Modified:
		eventTypeS = pretty.Yellow(string(eventType))
	case k8swatch.Error:
		eventTypeS = pretty.Red(string(eventType))
	default:
		eventTypeS = pretty.Cyan(string(eventType))
	}
	fmt.Fprintf(w, "    [%s] %s\n", eventTypeS, row.Message)
}

// TODO: refactor me
func hasDependents(res *models.Resource) (*models.Resource, bool) {
	// Parse nested field: kind
	kind, ok, _ := unstructured.NestedString(res.Attributes, "kind")
	if !ok {
		return nil, false
	}
	switch kind {
	case k8s.Service:
		// Endpoints will only be created if the service's selector is not nil
		selector, _, _ := unstructured.NestedMap(res.Attributes, "spec", "selector")
		if len(selector) == 0 {
			return nil, false
		}

		subKind := k8s.Endpoints
		apiVersion, _, _ := unstructured.NestedString(res.Attributes, "apiVersion")
		namespace, _, _ := unstructured.NestedString(res.Attributes, "metadata", "namespace")
		name, _, _ := unstructured.NestedString(res.Attributes, "metadata", "name")
		ret := &models.Resource{
			ID: engine.BuildIDForKubernetes(apiVersion, subKind, namespace, name),
			Attributes: map[string]interface{}{
				"apiVersion": apiVersion,
				"kind":       subKind,
				"metadata": map[string]string{
					"namespace": namespace,
					"name":      name,
				},
			},
		}
		return ret, true
	default:
		return nil, false
	}
}

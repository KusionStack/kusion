package printers

import (
	k8swatch "k8s.io/apimachinery/pkg/watch"

	"kusionstack.io/kusion/pkg/util/pretty"
)

type Table struct {
	IDs  []string
	Rows map[string]*Row
}

type Row struct {
	Type   k8swatch.EventType
	Kind   string
	Name   string
	Detail string
}

func NewTable() *Table {
	return &Table{
		IDs:  []string{},
		Rows: map[string]*Row{},
	}
}

func NewRow(t k8swatch.EventType, kind, name, detail string) *Row {
	return &Row{
		Type:   t,
		Kind:   kind,
		Name:   name,
		Detail: detail,
	}
}

const READY k8swatch.EventType = "READY"

func (t *Table) InsertOrUpdate(id string, row *Row) {
	_, ok := t.Rows[id]
	if !ok {
		t.IDs = append(t.IDs, id)
	}
	t.Rows[id] = row
}

func (t *Table) IsCompleted() bool {
	for _, row := range t.Rows {
		if row.Type != READY {
			return false
		}
	}
	return true
}

func (t *Table) Print() [][]string {
	data := [][]string{}
	data = append(data, []string{"Type", "Kind", "Name", "Detail"})
	for _, id := range t.IDs {
		row := t.Rows[id]
		eventType := row.Type

		// Colored type
		eventTypeS := ""
		switch eventType {
		case k8swatch.Added:
			eventTypeS = pretty.Cyan(string(eventType))
		case k8swatch.Deleted:
			eventTypeS = pretty.Red(string(eventType))
		case k8swatch.Modified:
			eventTypeS = pretty.Yellow(string(eventType))
		case k8swatch.Error:
			eventTypeS = pretty.Red(string(eventType))
		default:
			eventTypeS = pretty.Green(string(eventType))
		}

		data = append(data, []string{eventTypeS, row.Kind, row.Name, row.Detail})
	}
	return data
}

package models

import (
	"encoding/json"

	"kusionstack.io/kusion/pkg/util/pretty"
)

// ActionType represents the kind of operation performed by a plan.  It evaluates to its string label.
type ActionType int64

// ActionType values
const (
	Undefined ActionType = iota // invalidate value
	UnChange                    // nothing to do.
	Create                      // creating a new resource.
	Update                      // updating an existing resource.
	Delete                      // deleting an existing resource.
)

func (t ActionType) String() string {
	return []string{
		"Undefined",
		"UnChange",
		"Create",
		"Update",
		"Delete",
	}[t]
}

func (t ActionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t ActionType) Ing() string {
	switch t {
	case Create:
		return "Creating"
	case Update:
		return "Updating"
	case Delete:
		return "Deleting"
	default:
		return "Unchanged"
	}
}

func (t ActionType) PrettyString() string {
	switch t {
	case UnChange:
		return pretty.Gray(t.Ing())
	case Create:
		return pretty.Green(t.Ing())
	case Update:
		return pretty.Blue(t.Ing())
	case Delete:
		return pretty.Red(t.Ing())
	default:
		return pretty.Normal(t.Ing())
	}
}

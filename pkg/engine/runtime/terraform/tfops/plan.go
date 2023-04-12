package tfops

import "encoding/json"

// PlanRepresentation is the top-level representation of the json format of a plan. It includes
// the complete config and current state.
// Ref: https://developer.hashicorp.com/terraform/internals/json-format#plan-representation
type PlanRepresentation struct {
	FormatVersion    string      `json:"format_version,omitempty"`
	TerraformVersion string      `json:"terraform_version,omitempty"`
	Variables        variables   `json:"variables,omitempty"`
	PlannedValues    stateValues `json:"planned_values,omitempty"`
	// ResourceDrift and ResourceChanges are sorted in a user-friendly order
	// that is undefined at this time, but consistent.
	ResourceDrift      []ResourceChange  `json:"resource_drift,omitempty"`
	ResourceChanges    []ResourceChange  `json:"resource_changes,omitempty"`
	OutputChanges      map[string]Change `json:"output_changes,omitempty"`
	PriorState         json.RawMessage   `json:"prior_state,omitempty"`
	Config             json.RawMessage   `json:"configuration,omitempty"`
	RelevantAttributes []ResourceAttr    `json:"relevant_attributes,omitempty"`
	Checks             json.RawMessage   `json:"checks,omitempty"`
}

// variables is the JSON representation of the variables provided to the current
// plan.
type variables map[string]*variable

type variable struct {
	Value json.RawMessage `json:"value,omitempty"`
}

// ResourceChange is a description of an individual change action that Terraform
// plans to use to move from the prior state to a new state matching the
// configuration.
type ResourceChange struct {
	// Address is the absolute resource address
	Address string `json:"address,omitempty"`

	// PreviousAddress is the absolute address that this resource instance had
	// at the conclusion of a previous run.
	//
	// This will typically be omitted, but will be present if the previous
	// resource instance was subject to a "moved" block that we handled in the
	// process of creating this plan.
	//
	// Note that this behavior diverges from the internal plan data structure,
	// where the previous address is set equal to the current address in the
	// common case, rather than being omitted.
	PreviousAddress string `json:"previous_address,omitempty"`

	// ModuleAddress is the module portion of the above address. Omitted if the
	// instance is in the root module.
	ModuleAddress string `json:"module_address,omitempty"`

	// "managed" or "data"
	Mode string `json:"mode,omitempty"`

	Type         string          `json:"type,omitempty"`
	Name         string          `json:"name,omitempty"`
	Index        json.RawMessage `json:"index,omitempty"`
	ProviderName string          `json:"provider_name,omitempty"`

	// "deposed", if set, indicates that this action applies to a "deposed"
	// object of the given instance rather than to its "current" object. Omitted
	// for changes to the current object.
	Deposed string `json:"deposed,omitempty"`

	// Change describes the change that will be made to this object
	Change Change `json:"change,omitempty"`

	// ActionReason is a keyword representing some optional extra context
	// for why the actions in Change.Actions were chosen.
	//
	// This extra detail is only for display purposes, to help a UI layer
	// present some additional explanation to a human user. The possible
	// values here might grow and change over time, so any consumer of this
	// information should be resilient to encountering unrecognized values
	// and treat them as an unspecified reason.
	ActionReason string `json:"action_reason,omitempty"`
}

// Change is the representation of a proposed change for an object.
type Change struct {
	// Actions are the actions that will be taken on the object selected by the
	// properties below. Valid actions values are:
	//    ["no-op"]
	//    ["create"]
	//    ["read"]
	//    ["update"]
	//    ["delete", "create"]
	//    ["create", "delete"]
	//    ["delete"]
	// The two "replace" actions are represented in this way to allow callers to
	// e.g. just scan the list for "delete" to recognize all three situations
	// where the object will be deleted, allowing for any new deletion
	// combinations that might be added in future.
	Actions []string `json:"actions,omitempty"`

	// Before and After are representations of the object value both before and
	// after the action. For ["create"] and ["delete"] actions, either "before"
	// or "after" is unset (respectively). For ["no-op"], the before and after
	// values are identical. The "after" value will be incomplete if there are
	// values within it that won't be known until after apply.
	Before json.RawMessage `json:"before,omitempty"`
	After  json.RawMessage `json:"after,omitempty"`

	// AfterUnknown is an object value with similar structure to After, but
	// with all unknown leaf values replaced with true, and all known leaf
	// values omitted.  This can be combined with After to reconstruct a full
	// value after the action, including values which will only be known after
	// apply.
	AfterUnknown json.RawMessage `json:"after_unknown,omitempty"`

	// BeforeSensitive and AfterSensitive are object values with similar
	// structure to Before and After, but with all sensitive leaf values
	// replaced with true, and all non-sensitive leaf values omitted. These
	// objects should be combined with Before and After to prevent accidental
	// display of sensitive values in user interfaces.
	BeforeSensitive json.RawMessage `json:"before_sensitive,omitempty"`
	AfterSensitive  json.RawMessage `json:"after_sensitive,omitempty"`

	// ReplacePaths is an array of arrays representing a set of paths into the
	// object value which resulted in the action being "replace". This will be
	// omitted if the action is not replace, or if no paths caused the
	// replacement (for example, if the resource was tainted). Each path
	// consists of one or more steps, each of which will be a number or a
	// string.
	ReplacePaths json.RawMessage `json:"replace_paths,omitempty"`
}

// ResourceAttr contains the address and attribute of an external for the
// RelevantAttributes in the plan.
type ResourceAttr struct {
	Resource string          `json:"resource"`
	Attr     json.RawMessage `json:"attribute"`
}

package tfops

import (
	"encoding/json"

	"github.com/zclconf/go-cty/cty"

	"kusionstack.io/kusion/pkg/engine/models"
)

// Terraform State schema from https://github.com/hashicorp/terraform/blob/main/internal/command/jsonstate/state.go
type TFState struct {
	FormatVersion    string       `json:"format_version,omitempty"`
	TerraformVersion string       `json:"terraform_version,omitempty"`
	Values           *stateValues `json:"values,omitempty"`
}

// stateValues is the common representation of resolved values for both the prior
// state (which is always complete) and the planned new state.
type stateValues struct {
	Outputs    map[string]output `json:"outputs,omitempty"`
	RootModule module            `json:"root_module,omitempty"`
}

type output struct {
	Sensitive bool            `json:"sensitive"`
	Value     json.RawMessage `json:"value,omitempty"`
	Type      json.RawMessage `json:"type,omitempty"`
}

// InstanceKey represents the key of an instance within an object that
// contains multiple instances due to using "count" or "for_each" arguments
// in configuration.
//
// IntKey and StringKey are the two implementations of this type. No other
// implementations are allowed. The single instance of an object that _isn't_
// using "count" or "for_each" is represented by NoKey, which is a nil
// InstanceKey.
type InstanceKey interface {
	instanceKeySigil()
	String() string

	// Value returns the cty.Value of the appropriate type for the InstanceKey
	// value.
	Value() cty.Value
}

// module is the representation of a module in state. This can be the root module
// or a child module
type module struct {
	// Resources are sorted in a user-friendly order that is undefined at this
	// time, but consistent.
	Resources []resource `json:"resources,omitempty"`

	// Address is the absolute module address, omitted for the root module
	Address string `json:"address,omitempty"`

	// Each module object can optionally have its own nested "child_modules",
	// recursively describing the full module tree.
	ChildModules []module `json:"child_modules,omitempty"`
}

// Resource is the representation of a resource in the state.
type resource struct {
	// Address is the absolute resource address
	Address string `json:"address,omitempty"`

	// Mode can be "managed" or "data"
	Mode string `json:"mode,omitempty"`

	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`

	// Index is omitted for a resource not using `count` or `for_each`.
	Index InstanceKey `json:"index,omitempty"`

	// ProviderName allows the property "type" to be interpreted unambiguously
	// in the unusual situation where a provider offers a resource type whose
	// name does not start with its own name, such as the "googlebeta" provider
	// offering "google_compute_instance".
	ProviderName string `json:"provider_name"`

	// SchemaVersion indicates which version of the resource type schema the
	// "values" property conforms to.
	SchemaVersion uint64 `json:"schema_version"`

	// AttributeValues is the JSON representation of the attribute values of the
	// resource, whose structure depends on the resource type schema. Any
	// unknown values are omitted or set to null, making them indistinguishable
	// from absent values.
	AttributeValues attributeValues `json:"values,omitempty"`

	// DependsOn contains a list of the resource's dependencies. The entries are
	// addresses relative to the containing module.
	DependsOn []string `json:"depends_on,omitempty"`

	// Tainted is true if the resource is tainted in terraform state.
	Tainted bool `json:"tainted,omitempty"`

	// Deposed is set if the resource is deposed in terraform state.
	DeposedKey string `json:"deposed_key,omitempty"`
}

// attributeValues is the JSON representation of the attribute values of the
// resource, whose structure depends on the resource type schema.
type attributeValues map[string]interface{}

// ConvertTFState convert Terraform State to kusion State
func ConvertTFState(tfState *TFState, providerAddr string) models.Resource {
	if tfState == nil || tfState.Values == nil {
		return models.Resource{}
	}
	// terraform runtime execute single node
	tResource := tfState.Values.RootModule.Resources[0]
	extension := make(map[string]interface{})
	extension["resourceType"] = tResource.Type
	extension["provider"] = providerAddr
	r := models.Resource{
		ID:         tResource.Name,
		Type:       "Terraform",
		Attributes: tResource.AttributeValues,
		Extensions: extension,
	}

	return r
}

package scaffold

import (
	"fmt"
	"strconv"
)

// ProjectTemplate is a Kusion project template manifest.
type ProjectTemplate struct {
	// ProjectName is a required fully qualified name.
	ProjectName string `json:"projectName" yaml:"projectName"`
	// Description is an optional description of the template.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Quickstart contains optional text to be displayed after template creation.
	Quickstart string `json:"quickstart,omitempty" yaml:"quickstart,omitempty"`
	// ProjectFields contains configuration in project level
	ProjectFields []*FieldTemplate `json:"projectFields,omitempty" yaml:"projectFields,omitempty"`
	// StackTemplates contains configuration in stack level
	StackTemplates []*StackTemplate `json:"stacks,omitempty" yaml:"stacks,omitempty"`
}

// StackTemplate contains configuration in stack level
type StackTemplate struct {
	// Name is stack name
	Name string `json:"name" yaml:"name"`
	// Fields contains all stack fields definition
	Fields []*FieldTemplate `json:"fields,omitempty" yaml:"fields,omitempty"`
}

// FieldTemplate can describe all kinds of type, including primitive and composite.
type FieldTemplate struct {
	// Name represents the field name, required
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Description represents the field description, optional
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Type can be string/int/bool/float/array/map/struct/any, required
	Type FieldType `json:"type,omitempty" yaml:"type,omitempty"`
	// Default represents default value for all FieldType
	Default interface{} `json:"default,omitempty" yaml:"default,omitempty"`
	// Elem is effective only when type is ArrayField
	Elem *FieldTemplate `json:"elem,omitempty" yaml:"elem,omitempty"`
	// Key is effective only when type is MapField
	Key *FieldTemplate `json:"key,omitempty" yaml:"key,omitempty"`
	// Value is effective only when type is MapField
	Value *FieldTemplate `json:"value,omitempty" yaml:"value,omitempty"`
	// Fields is effective only when type is StructField
	Fields []*FieldTemplate `json:"fields,omitempty" yaml:"fields,omitempty"`
}

// FieldType includes field type that can be unmarshalled directly
type FieldType string

const (
	StringField FieldType = "string"
	IntField    FieldType = "int"
	BoolField   FieldType = "bool"
	FloatField  FieldType = "float"
	ArrayField  FieldType = "array"
	MapField    FieldType = "map"
	StructField FieldType = "struct"
	AnyField    FieldType = "any" // AnyField equal to interface{}
)

// IsPrimitive indicate the give field is one of StringField, IntField, FloatField, BoolField or not
func (f FieldType) IsPrimitive() bool {
	return f == StringField || f == IntField || f == FloatField || f == BoolField
}

// RestoreActualValue help to transfer input to actual value according to its type
func (f *FieldTemplate) RestoreActualValue(input string) (actual interface{}, err error) {
	if f.Type == "" {
		return nil, fmt.Errorf("field %s miss type definition", f.Name)
	}
	switch f.Type {
	case IntField:
		actual, err = strconv.Atoi(input)
	case BoolField:
		actual, err = strconv.ParseBool(input)
	case FloatField:
		actual, err = strconv.ParseFloat(input, 64)
	case StringField:
		return input, nil
	}
	return actual, err
}

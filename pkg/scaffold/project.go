package scaffold

import "strconv"

// ProjectTemplate is a Kusion project template manifest.
type ProjectTemplate struct {
	// ProjectName is a required fully qualified name.
	ProjectName string `json:"projectName" yaml:"projectName"`
	// Description is an optional description of the template.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Quickstart contains optional text to be displayed after template creation.
	Quickstart string `json:"quickstart,omitempty" yaml:"quickstart,omitempty"`
	// CommonTemplates contains configuration in project level
	CommonTemplates []*FieldTemplate `json:"common,omitempty" yaml:"common,omitempty"`
	// StackTemplates contains configuration in stack level
	StackTemplates []*StackTemplate `json:"stacks,omitempty" yaml:"stacks,omitempty"`
}

type StackTemplate struct {
	// Name is stack name
	Name string `json:"name" yaml:"name"`
	// Fields contains all fields wait to be prompt
	Fields []*FieldTemplate `json:"fields,omitempty" yaml:"fields,omitempty"`
}

type FieldTemplate struct {
	// Name represents the field name, required
	Name string `json:"name" yaml:"name"`
	// Description represents the field description, optional
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Type can be string/int/bool/float/array/struct, required
	Type FieldType `json:"type" yaml:"type"`
	// Default represents primitive field default value
	Default interface{} `json:"default,omitempty" yaml:"default,omitempty"`
	// Elem is active only when type is ArrayField
	Elem *FieldTemplate `json:"elem,omitempty" yaml:"elem,omitempty"`
	// Fields is active only when type is StructField
	Fields []*FieldTemplate `json:"fields,omitempty" yaml:"fields,omitempty"`
}

type FieldType string

const (
	StringField FieldType = "string"
	IntField    FieldType = "int"
	BoolField   FieldType = "bool"
	FloatField  FieldType = "float"
	ArrayField  FieldType = "array"
	StructField FieldType = "struct"
)

func (f FieldType) IsPrimitive() bool {
	return f == "" || f == StringField || f == IntField || f == FloatField || f == BoolField
}

func (f *FieldTemplate) RestoreActualValue(input string) (actual interface{}, err error) {
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

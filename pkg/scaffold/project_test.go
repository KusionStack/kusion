package scaffold

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldTemplate_RestoreActualValue(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		Type        FieldType
		Default     interface{}
		Elem        *FieldTemplate
		Key         *FieldTemplate
		Value       *FieldTemplate
		Fields      []*FieldTemplate
	}
	type args struct {
		input string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantActual interface{}
		wantErr    bool
	}{
		{
			name: "string",
			fields: fields{
				Type: StringField,
			},
			args: args{
				input: "foo",
			},
			wantActual: "foo",
			wantErr:    false,
		},
		{
			name: "array",
			fields: fields{
				Type: ArrayField,
				Elem: &FieldTemplate{
					Type: IntField,
				},
				Default: []int{1, 2, 3},
			},
			wantActual: nil,
			wantErr:    false,
		},
		{
			name: "map",
			fields: fields{
				Type: MapField,
				Key: &FieldTemplate{
					Type: StringField,
				},
				Value: &FieldTemplate{
					Type: BoolField,
				},
				Default: map[string]bool{
					"foo": true,
					"bar": false,
				},
			},
			wantActual: nil,
			wantErr:    false,
		},
		{
			name: "struct",
			fields: fields{
				Type: StructField,
				Fields: []*FieldTemplate{
					{
						Name: "float field",
						Type: FloatField,
					},
					{
						Name: "array field",
						Type: ArrayField,
						Elem: &FieldTemplate{
							Type: IntField,
						},
					},
					{
						Name: "map field",
						Type: MapField,
						Key: &FieldTemplate{
							Type: StringField,
						},
						Value: &FieldTemplate{
							Type: BoolField,
						},
					},
					{
						Name: "inner struct",
						Type: StructField,
						Fields: []*FieldTemplate{
							{
								Name: "foo",
								Type: StringField,
							},
						},
					},
				},
				Default: map[string]interface{}{
					"string field": "foo",
					"array field":  []int{1, 2, 3},
					"map field": map[string]bool{
						"foo": true,
						"bar": false,
					},
					"inner struct": map[string]interface{}{
						"foo": "bar",
					},
				},
			},
			wantActual: nil,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FieldTemplate{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Type:        tt.fields.Type,
				Default:     tt.fields.Default,
				Elem:        tt.fields.Elem,
				Key:         tt.fields.Key,
				Value:       tt.fields.Value,
				Fields:      tt.fields.Fields,
			}
			gotActual, err := f.RestoreActualValue(tt.args.input)
			assert.Equalf(t, tt.wantActual, gotActual, "RestoreActualValue(%v)", tt.args.input)
			assert.Equalf(t, tt.wantErr, err != nil, "RestoreActualValue(%v), err: %v", tt.args.input, err)
		})
	}
}

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
			name: "bool",
			fields: fields{
				Type: BoolField,
			},
			args: args{
				input: "true",
			},
			wantActual: true,
			wantErr:    false,
		},
		{
			name: "int",
			fields: fields{
				Type: IntField,
			},
			args: args{
				input: "1024",
			},
			wantActual: 1024,
			wantErr:    false,
		},
		{
			name: "float",
			fields: fields{
				Type: FloatField,
			},
			args: args{
				input: "3.1415926",
			},
			wantActual: 3.1415926,
			wantErr:    false,
		},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FieldTemplate{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Type:        tt.fields.Type,
				Default:     tt.fields.Default,
				Elem:        tt.fields.Elem,
				Fields:      tt.fields.Fields,
			}
			gotActual, err := f.RestoreActualValue(tt.args.input)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equalf(t, tt.wantActual, gotActual, "RestoreActualValue(%v)", tt.args.input)
		})
	}
}

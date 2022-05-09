package json

import (
	"reflect"
	"testing"
)

var mockConfigs = []interface{}{
	map[string]interface{}{
		"key1": "value1",
	},
	[]interface{}{
		"value2",
	},
	"value3",
}

type TestStruct struct {
	Name string `json:"name"`
}

func TestRemoveListFields(t *testing.T) {
	type args struct {
		config []interface{}
		live   []interface{}
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "success",
			args: args{
				config: mockConfigs,
				live:   mockConfigs,
			},
			want: mockConfigs,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveListFields(tt.args.config, tt.args.live); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveListFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarshal2String(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				v: TestStruct{
					Name: "test",
				},
			},
			want: `{"name":"test"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Marshal2String(tt.args.v); got != tt.want {
				t.Errorf("Marshal2String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustMarshal2String(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				v: TestStruct{
					Name: "test",
				},
			},
			want: `{"name":"test"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MustMarshal2String(tt.args.v); got != tt.want {
				t.Errorf("MustMarshal2String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustMarshal2PrettyString(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				v: TestStruct{
					Name: "test",
				},
			},
			want: `{
  "name": "test"
}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MustMarshal2PrettyString(tt.args.v); got != tt.want {
				t.Errorf("MustMarshal2PrettyString() = %v, want %v", got, tt.want)
			}
		})
	}
}

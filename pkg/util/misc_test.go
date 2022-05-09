package util

import "testing"

func TestInArray(t *testing.T) {
	type args struct {
		x     string
		array []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty slice",
			args: args{
				x:     "a",
				array: []string{},
			},
			want: false,
		},
		{
			name: "in array",
			args: args{
				x:     "a",
				array: []string{"a", "b", "c"},
			},
			want: true,
		},
		{
			name: "not in array",
			args: args{
				x:     "d",
				array: []string{"a", "b", "c"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InArray(tt.args.x, tt.args.array); got != tt.want {
				t.Errorf("InArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsZero(t *testing.T) {
	type args struct {
		x interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "string zero",
			args: args{
				x: "",
			},
			want: true,
		},
		{
			name: "bool zero",
			args: args{
				x: false,
			},
			want: true,
		},
		{
			name: "int zero",
			args: args{
				x: 0,
			},
			want: true,
		},
		{
			name: "uint zero",
			args: args{
				x: uint(0),
			},
			want: true,
		},
		{
			name: "float zero",
			args: args{
				x: 0.0,
			},
			want: true,
		},
		{
			name: "ptr zero",
			args: args{
				x: new(string),
			},
			want: false,
		},
		{
			name: "nil",
			args: args{},
			want: true,
		},
		{
			name: "struct nil",
			args: args{
				x: struct {
					x string
				}{},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsZero(tt.args.x); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

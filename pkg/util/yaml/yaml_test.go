package yaml

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

var (
	mockNamespace    *ast.DocumentNode
	mockNamespaceMap map[string]interface{} = map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata": map[string]interface{}{
			"name": "default",
			"labels": map[string]interface{}{
				"kubernetes.io/metadata.name": "default",
			},
		},
	}
)

const (
	mockNamespaceKind   = "Namespace"
	mockNamespaceString = `apiVersion: v1
kind: Namespace
metadata:
  labels:
    kubernetes.io/metadata.name: default
  name: default
`
)

func builder() *yaml.PathBuilder {
	return &yaml.PathBuilder{}
}

func TestParseYamlFromFile(t *testing.T) {
	type args struct {
		filename string
		target   interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				filename: "./testdata/test-success.yaml",
				target:   map[string]interface{}{},
			},
			wantErr: false,
		},
		{
			name: "yaml unmarshal fail",
			args: args{
				filename: "./testdata/test-fail.yaml",
				target:   []string{},
			},
			wantErr: true,
		},
		{
			name: "read file fail",
			args: args{
				filename: "./testdata",
				target:   map[string]interface{}{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseYamlFromFile(tt.args.filename, tt.args.target); (err != nil) != tt.wantErr {
				t.Errorf("ParseYamlFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFile2Document(t *testing.T) {
	// Run test
	type args struct {
		filename   string
		filterKind string
	}
	tests := []struct {
		name string
		args args
		want *ast.DocumentNode
	}{
		{
			name: "filter success",
			args: args{
				filename:   "./testdata/namespace.yaml",
				filterKind: mockNamespaceKind,
			},
			want: mockNamespace,
		},
		{
			name: "filter fail",
			args: args{
				filename:   "./testdata/namespace.yaml",
				filterKind: "Secret",
			},
			want: nil,
		},
		{
			name: "filter empty",
			args: args{
				filename:   "./testdata/namespace.yaml",
				filterKind: "",
			},
			want: mockNamespace,
		},
		{
			name: "filter failed file",
			args: args{
				filename:   "./testdata/",
				filterKind: "",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := File2Document(tt.args.filename, tt.args.filterKind); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("File2Document() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYAML2Documents(t *testing.T) {
	type args struct {
		yamlContent string
	}
	tests := []struct {
		name    string
		args    args
		want    []*ast.DocumentNode
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				yamlContent: mockNamespaceString,
			},
			want:    []*ast.DocumentNode{mockNamespace},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := YAML2Documents(tt.args.yamlContent)
			if (err != nil) != tt.wantErr {
				t.Errorf("YAML2Documents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("YAML2Documents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetByPathString(t *testing.T) {
	type args struct {
		doc  io.Reader
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				doc:  mockNamespace,
				path: "$.kind",
			},
			want:    mockNamespaceKind,
			wantErr: false,
		},
		{
			name: "fail",
			args: args{
				doc:  mockNamespace,
				path: "$.fail",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetByPathString(tt.args.doc, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByPathString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetByPathString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustGetByPathString(t *testing.T) {
	type args struct {
		doc  io.Reader
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				doc:  mockNamespace,
				path: "$.kind",
			},
			want: mockNamespaceKind,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MustGetByPathString(tt.args.doc, tt.args.path); got != tt.want {
				t.Errorf("MustGetByPathString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetByPath(t *testing.T) {
	type args struct {
		doc  io.Reader
		path *yaml.Path
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				doc:  mockNamespace,
				path: builder().Root().Child("kind").Build(),
			},
			want:    mockNamespaceKind,
			wantErr: false,
		},
		{
			name: "fail",
			args: args{
				doc:  mockNamespace,
				path: builder().Root().Child("fail").Build(),
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetByPath(tt.args.doc, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetByPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustGetByPath(t *testing.T) {
	type args struct {
		doc  io.Reader
		path *yaml.Path
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				doc:  mockNamespace,
				path: builder().Root().Child("kind").Build(),
			},
			want: mockNamespaceKind,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MustGetByPath(tt.args.doc, tt.args.path); got != tt.want {
				t.Errorf("MustGetByPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeToOneYAML(t *testing.T) {
	type args struct {
		yamlList []interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				yamlList: []interface{}{mockNamespaceMap},
			},
			want: mockNamespaceString,
		},
		{
			name: "empty",
			args: args{
				yamlList: []interface{}{},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeToOneYAML(tt.args.yamlList...); got != tt.want {
				t.Errorf("MergeToOneYAML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeStringsToOneYAML(t *testing.T) {
	type args struct {
		yamlList []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				yamlList: []string{mockNamespaceString, mockNamespaceString},
			},
			want: fmt.Sprintf("%s---\n%s", mockNamespaceString, mockNamespaceString),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeStringsToOneYAML(tt.args.yamlList); got != tt.want {
				t.Errorf("MergeStringsToOneYAML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMain(t *testing.M) {
	// Pre operation
	namespaces, err := YAML2Documents(mockNamespaceString)
	if err != nil {
		panic(err)
	}
	mockNamespace = namespaces[0]

	// Run test
	os.Exit(t.Run())
}

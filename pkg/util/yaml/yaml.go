package yaml

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	yamlv3 "gopkg.in/yaml.v3"
)

// Parse yaml data by file name
func ParseYamlFromFile(filename string, target interface{}) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yamlv3.Unmarshal(content, target)
	if err != nil {
		return err
	}
	return nil
}

// Read specify document by filterKind from specify file,
// return first document when filterKind is empty
func File2Document(filename string, filterKind string) *ast.DocumentNode {
	file, err := parser.ParseFile(filename, 0)
	if err != nil || len(file.Docs) == 0 {
		return nil
	}
	if filterKind == "" {
		return file.Docs[0]
	} else {
		for _, doc := range file.Docs {
			docKind, err := GetByPathString(doc, "$.kind")
			if err != nil {
				continue
			}
			if docKind == filterKind {
				return doc
			}
		}
		return nil
	}
}

// Convert yaml string to []ast.Document
func YAML2Documents(yamlContent string) ([]*ast.DocumentNode, error) {
	// parser documents from yaml
	file, err := parser.ParseBytes([]byte(yamlContent), 0)
	if err != nil {
		return nil, err
	}
	return file.Docs, nil
}

// Get specify node string from yaml by path string
func GetByPathString(doc io.Reader, path string) (string, error) {
	var result string
	p, err := yaml.PathString(path)
	if err != nil {
		return "", fmt.Errorf("unable build path by %s: %v", path, err)
	}
	err = p.Read(doc, &result)
	if err != nil {
		return "", fmt.Errorf("unable read by path %s: %v", path, err)
	}
	return result, nil
}

// Get specify node string from yaml by path string, panic if occur error
func MustGetByPathString(doc io.Reader, path string) string {
	var result string
	p, err := yaml.PathString(path)
	if err != nil {
		panic(fmt.Sprintf("unable build path by %s: %v", path, err))
	}
	err = p.Read(doc, &result)
	if err != nil {
		panic(fmt.Sprintf("unable build path by %s: %v", path, err))
	}
	return result
}

// Get specify node string from yaml by path, panic if occur error
func GetByPath(doc io.Reader, path *yaml.Path) (string, error) {
	expectNode, err := path.ReadNode(doc)
	if err != nil {
		return "", fmt.Errorf("unable read by path %s: %v", path, err)
	}
	if stringNode, ok := expectNode.(*ast.StringNode); ok {
		return stringNode.Value, nil
	}
	return expectNode.String(), nil
}

// Get specify node string from yaml by path, panic if occur error
func MustGetByPath(doc io.Reader, path *yaml.Path) string {
	expectNode, err := path.ReadNode(doc)
	if err != nil {
		panic(fmt.Sprintf("unable get value by path: %v", path))
	}
	if stringNode, ok := expectNode.(*ast.StringNode); ok {
		return stringNode.Value
	}
	return expectNode.String()
}

// TODO: yamlv3.Marshal will reduce leading "/n" character
// Merge multiple yaml documents into a single string,
// separate yaml documents with '---'
func MergeToOneYAML(yamlList ...interface{}) string {
	if len(yamlList) == 0 {
		return ""
	}
	var result bytes.Buffer
	e := yamlv3.NewEncoder(&result)
	for _, y := range yamlList {
		// compatible with basic type
		if y == nil || reflect.ValueOf(y).IsZero() {
			y = nil
		}
		e.SetIndent(2)
		err := e.Encode(y)
		if err != nil {
			panic(err)
		}
	}
	err := e.Close()
	if err != nil {
		panic(err)
	}
	return result.String()
}

// Merge multiple yaml string into a single string,
// separate yaml documents with '---'
func MergeStringsToOneYAML(yamlList []string) string {
	documents := []interface{}{}
	for _, y := range yamlList {
		document := map[string]interface{}{}
		yamlv3.Unmarshal([]byte(y), document)
		documents = append(documents, document)
	}
	return MergeToOneYAML(documents...)
}

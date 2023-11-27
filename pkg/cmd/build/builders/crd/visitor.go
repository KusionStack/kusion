package crd

// Visitor walks a list of resources under target path.
type Visitor interface {
	Visit() ([]interface{}, error)
}

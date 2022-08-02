package resources

// Visitor walks a list of resources under target path.
type Visitor interface {
	Visit() ([]interface{}, error)
}

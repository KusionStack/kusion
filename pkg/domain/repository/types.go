package repository

// Bound represents the query bound for a database access.
type Bound struct {
	// Offset is the number of items to skip.
	Offset int
	// Limit is the maximum number of items to return.
	Limit int
}

// Condition represents the query conditions for a database access.
type Condition struct {
	// Keyword is the keyword to search for.
	Keyword string
}

// Query represents the query criteria for a database access.
type Query struct {
	Bound
	Condition
}

// StackCondition represents the stack query conditions for a database
// access.
type StackCondition struct {
	Condition
	SourceIDs []string
	Desired   string
	Framework string
	State     string
}

// SourceCondition represents the source query conditions for a database access.
type SourceCondition struct {
	Condition
	SourceProvider string
}

// StackQuery represents the stack query criteria for a database access.
type StackQuery struct {
	Bound
	StackCondition
}

// SourceQuery represents the source query criteria for a database access.
type SourceQuery struct {
	Bound
	SourceCondition
}

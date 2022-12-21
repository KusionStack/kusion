package models

type OperationType int64

// Operation type
const (
	UndefinedOperation OperationType = iota // invalidate value
	Apply
	ApplyPreview
	Destroy
	DestroyPreview
)

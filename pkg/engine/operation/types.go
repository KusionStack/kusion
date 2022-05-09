package operation

type Type int64

// Operation type
const (
	UndefinedOperation Type = iota // invalidate value
	Apply
	Preview
	Destroy
)

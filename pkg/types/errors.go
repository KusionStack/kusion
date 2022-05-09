package types

type UnsupportedErr struct{}

func (e *UnsupportedErr) Error() string {
	return "unsupported"
}

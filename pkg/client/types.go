package client

import "errors"

// Generic api errors.
// Errors returned by api can be tested against these errors
// using errors.Is.
var (
	ErrResponseFailed = errors.New("the response failed")
)

package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

var (
	ErrProjectDoesNotExist      = errors.New("the project does not exist")
	ErrOrganizationDoesNotExist = errors.New("the organization does not exist")
	ErrStackDoesNotExist        = errors.New("the stack does not exist")
)

// Payload is an interface for incoming requests payloads
// Each handler should implement this interface to parse payloads
type Payload interface {
	Decode(*http.Request) error // Decode returns the payload object with the decoded
}

// response defines the structure for API response payloads.
type Response struct {
	Success   bool       `json:"success" yaml:"success"`                         // Indicates success status.
	Message   string     `json:"message" yaml:"message"`                         // Descriptive message.
	Data      any        `json:"data,omitempty" yaml:"data,omitempty"`           // Data payload.
	TraceID   string     `json:"traceID,omitempty" yaml:"traceID,omitempty"`     // Trace identifier.
	StartTime *time.Time `json:"startTime,omitempty" yaml:"startTime,omitempty"` // Request start time.
	EndTime   *time.Time `json:"endTime,omitempty" yaml:"endTime,omitempty"`     // Request end time.
	CostTime  Duration   `json:"costTime,omitempty" yaml:"costTime,omitempty"`   // Time taken for the request.
}

// Render is a no-op method that satisfies the render.Renderer interface.
func (rep *Response) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Duration is a custom type that represents a duration of time.
type Duration time.Duration

// MarshalJSON customizes JSON representation of the Duration type.
func (d Duration) MarshalJSON() (b []byte, err error) {
	// Format the duration as a string.
	return []byte(fmt.Sprintf(`"%s"`, time.Duration(d).String())), nil
}

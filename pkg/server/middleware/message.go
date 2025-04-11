package middleware

import (
	"context"
)

// Key to use when setting the response message.
type (
	ctxKeyResponseMessage string
)

// ResponseMessageKey is a context key used for associating a message with a response.
const ResponseMessageKey ctxKeyResponseMessage = "response_message"

// GetResponseMessage returns a response message from the given context.
func GetResponseMessage(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if responseMessage, ok := ctx.Value(ResponseMessageKey).(string); ok {
		return responseMessage
	}

	return ""
}

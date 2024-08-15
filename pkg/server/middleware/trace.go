package middleware

// Ported from Goji's middleware, source:
// https://github.com/zenazn/goji/tree/master/web/middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/go-chi/httplog/v2"
	"kusionstack.io/kusion/pkg/domain/constant"
)

// Key to use when setting the trace ID and user ID.
type ctxKeyTraceID int
type ctxKeyUserID string

// TraceIDKey and UserIDKey are the keys that hold the unique
// trace ID and user ID in a request context.
const TraceIDKey ctxKeyTraceID = 0
const UserIDKey ctxKeyUserID = "user_id"

// TraceIDHeader and UserIDHeader is the name of the HTTP Header
// which contains the trace id and user id.
var TraceIDHeader = "x-kusion-trace"
var UserIDHeader = "x-kusion-user"

var prefix string
var reqid uint64

// A quick note on the statistics here: we're trying to calculate the chance that
// two randomly generated base62 prefixes will collide. We use the formula from
// http://en.wikipedia.org/wiki/Birthday_problem
//
// P[m, n] \approx 1 - e^{-m^2/2n}
//
// We ballpark an upper bound for $m$ by imagining (for whatever reason) a server
// that restarts every second over 10 years, for $m = 86400 * 365 * 10 = 315360000$
//
// For a $k$ character base-62 identifier, we have $n(k) = 62^k$
//
// Plugging this in, we find $P[m, n(10)] \approx 5.75%$, which is good enough for
// our purposes, and is surely more than anyone would ever need in practice -- a
// process that is rebooted a handful of times a day for a hundred years has less
// than a millionth of a percent chance of generating two colliding IDs.

func init() {
	hostname, err := os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}
	var buf [12]byte
	var b64 string
	for len(b64) < 10 {
		rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}

	prefix = fmt.Sprintf("%s/%s", hostname, b64[0:10])
}

// TraceID is a middleware that injects a trace ID into the context of each
// request. A trace ID is a string of the form "host.example.com/random-0001",
// where "random" is a base62 random string that uniquely identifies this go
// process, and where the last number is an atomically incremented request
// counter.
func TraceID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		traceID := r.Header.Get(TraceIDHeader)
		if traceID == "" {
			myid := atomic.AddUint64(&reqid, 1)
			traceID = fmt.Sprintf("%s-%06d", prefix, myid)
		}
		ctx = context.WithValue(ctx, TraceIDKey, traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// GetTraceID returns a trace ID from the given context if one is present.
// Returns the empty string if a trace ID cannot be found.
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// NextTraceID generates the next trace ID in the sequence.
func NextTraceID() uint64 {
	return atomic.AddUint64(&reqid, 1)
}

// UserID is a middleware that injects the operator of the request
// into the context of each request.
func UserID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := r.Header.Get(UserIDHeader)
		if userID == "" {
			userID = constant.DefaultUser
		}
		ctx = context.WithValue(ctx, UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// GetUserID returns a user ID from the given context if one is present.
// Returns the empty string if a user ID cannot be found.
func GetUserID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	var logger *httplog.Logger
	if apiLogger, ok := ctx.Value(APILoggerKey).(*httplog.Logger); ok {
		logger = apiLogger
	} else {
		logger = httplog.NewLogger("DefaultLogger")
	}

	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		logger.Info("User ID: ", "user_id", userID)
		return userID
	}
	return ""
}

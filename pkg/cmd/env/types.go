// Reference: https://github.com/golang/go/blob/master/src/cmd/go/internal/cfg/cfg.go
package env

// An EnvVar is an environment variable Name=Value.
type EnvVar struct {
	Name  string
	Value string
}

package init

import (
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/engine/states/local"
	"kusionstack.io/kusion/pkg/engine/states/remote/db"
	"kusionstack.io/kusion/pkg/engine/states/remote/http"
	"kusionstack.io/kusion/pkg/engine/states/remote/oss"
	"kusionstack.io/kusion/pkg/engine/states/remote/s3"
)

// backends store all available backend
var backends map[string]func() states.Backend

// init backends map with all support backend
func init() {
	backends = map[string]func() states.Backend{
		"local": local.NewLocalBackend,
		"db":    db.NewDBBackend,
		"oss":   oss.NewOssBackend,
		"s3":    s3.NewS3Backend,
		"http":  http.NewHTTPBackend,
	}
}

// GetBackend return backend, or nil if not exists
func GetBackend(name string) func() states.Backend {
	return backends[name]
}

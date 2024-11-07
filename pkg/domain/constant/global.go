package constant

import "time"

// These constants represent the possible states of a stack.
const (
	DefaultUser        = "test.user"
	DefaultWorkspace   = "default"
	DefaultBackend     = "default"
	DefaultOrgOwner    = "kusion"
	DefaultSourceType  = SourceProviderTypeGit
	DefaultSourceDesc  = "Default source"
	DefaultSystemName  = "kusion"
	MaxConcurrent      = 10
	DefaultLogFilePath = "/home/admin/logs/kusion.log"
	RepoCacheTTL       = 60 * time.Minute
)

// Package registry is used to keep all the module, runtime, backend and the corresponding
// config structure, so that the workspace config can be correctly deserialized.
package registry

import (
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/accessories/database"
)

var (
	registeredModules  []models.Module
	registeredRuntimes []models.RuntimeConfig
	registeredBackends []models.BackendConfig
)

func init() {
	registeredModules = []models.Module{
		&database.Database{},
	}
}

func GetRegisteredModules() []models.Module {
	return registeredModules
}

func GetRegisteredRuntimes() []models.RuntimeConfig {
	return registeredRuntimes
}

func GetRegisteredBackends() []models.BackendConfig {
	return registeredBackends
}

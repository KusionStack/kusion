package models

// Module is an interface to describe the domain model of a building block. A Module comes both
// from an application developer-centric schema and workspace config, which can fully describes a workload,
// an accessory, etc., such as database.Database, network.Port.
// A Module can be built into a or a set of Resource, which will be consumed by Kusion engine.
type Module interface {
	// ModuleName returns the name to identify the module uniquely.
	ModuleName() string
}

// SecureWorkspaceModule is used to hold Module fields retrieved from workspace config which get
// verified.
type SecureWorkspaceModule interface {
	Module

	// ValidateWorkspaceModule validates the Module retrieved from workspace is correct or not.
	ValidateWorkspaceModule() error
}

// SecureSchemaModule is used to hold Module fields retrieved from application developer-centric schema
// which get verified.
type SecureSchemaModule interface {
	Module

	// ValidateSchemaModule validates the Module retrieved from schema is correct or not.
	ValidateSchemaModule() error
}

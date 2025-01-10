package constant

import "errors"

var (
	ErrProjectNil               = errors.New("project is nil")
	ErrProjectName              = errors.New("project must have a name")
	ErrProjectSource            = errors.New("project must have a source")
	ErrProjectOrganization      = errors.New("project must have an organization")
	ErrProjectSourceProvider    = errors.New("project source must have a source provider")
	ErrProjectRemote            = errors.New("project source must have a remote")
	ErrProjectCreationTimestamp = errors.New("project must have a creation timestamp")
	ErrProjectUpdateTimestamp   = errors.New("project must have a update timestamp")
	ErrProjectPath              = errors.New("project must have a path")
	ErrInvalidProjectPath       = errors.New("project path can only have alphanumeric characters, slashes and underscores with [\\/a-zA-Z0-9_]")
	ErrOrgIDOrDomainRequired    = errors.New("either domain or organization ID is required")
	ErrInvalidStackPath         = errors.New("stack path can only have alphanumeric characters, slashes and underscores with [\\/a-zA-Z0-9_]")
	ErrInvalidProjectName       = errors.New("project name can only have alphanumeric characters and underscores with [a-zA-Z0-9_]")
	ErrInvalidStackName         = errors.New("stack name can only have alphanumeric characters and underscores with [a-zA-Z0-9_]")
	ErrInvalidAppConfigName     = errors.New("appConfig name can only have alphanumeric characters and underscores with [a-zA-Z0-9_]")
	ErrInvalidProjectID         = errors.New("the project ID should be a uuid")
	ErrInvalidStackID           = errors.New("the stack ID should be a uuid")
	ErrProjectNameAndFuzzyName  = errors.New("project name and fuzzy name cannot be set at the same time")
)

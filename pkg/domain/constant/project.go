package constant

import "errors"

var (
	ErrProjectNil               = errors.New("project is nil")
	ErrProjectName              = errors.New("project must have a name")
	ErrProjectSource            = errors.New("project must have a source")
	ErrProjectSourceProvider    = errors.New("project source must have a source provider")
	ErrProjectRemote            = errors.New("project source must have a remote")
	ErrProjectCreationTimestamp = errors.New("project must have a creation timestamp")
	ErrProjectUpdateTimestamp   = errors.New("project must have a update timestamp")
	ErrProjectPath              = errors.New("project must have a path")
)

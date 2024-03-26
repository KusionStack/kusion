package sourceproviders

import (
	"errors"
)

var (
	ErrCheckingOutBranch     = errors.New("err checking out branch")
	ErrCheckingOutRevision   = errors.New("err checking out revision")
	ErrCheckingOutRepository = errors.New("err checking out repository")
)

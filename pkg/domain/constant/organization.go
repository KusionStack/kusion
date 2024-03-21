package constant

import "errors"

var (
	ErrOrgNil               = errors.New("organization is nil")
	ErrOrgNameEmpty         = errors.New("organization must have a name")
	ErrOrganizationOwnerNil = errors.New("org must have at least one owner")
)

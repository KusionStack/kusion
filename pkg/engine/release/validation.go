package release

import (
	"errors"
	"fmt"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

var (
	ErrEmptyRelease         = errors.New("empty release")
	ErrEmptyProject         = errors.New("empty project")
	ErrEmptyWorkspace       = errors.New("empty workspace")
	ErrEmptyRevision        = errors.New("empty revision")
	ErrEmptyStack           = errors.New("empty stack")
	ErrEmptySpec            = errors.New("empty spec")
	ErrEmptyState           = errors.New("empty state")
	ErrEmptyPhase           = errors.New("empty phase")
	ErrEmptyCreateTime      = errors.New("empty create time")
	ErrEmptyModifiedTime    = errors.New("empty modified time")
	ErrDuplicateResourceKey = errors.New("duplicate resource key")
)

func ValidateRelease(r *v1.Release) error {
	if r == nil {
		return ErrEmptyRelease
	}
	if r.Project == "" {
		return ErrEmptyProject
	}
	if r.Workspace == "" {
		return ErrEmptyWorkspace
	}
	if r.Revision == 0 {
		return ErrEmptyRevision
	}
	if r.Stack == "" {
		return ErrEmptyStack
	}
	if err := ValidateSpec(r.Spec); err != nil {
		return err
	}
	if err := validateState(r.State); err != nil {
		return err
	}
	if r.Phase == "" {
		return ErrEmptyPhase
	}
	if r.CreateTime.IsZero() {
		return ErrEmptyCreateTime
	}
	if r.ModifiedTime.IsZero() {
		return ErrEmptyModifiedTime
	}
	return nil
}

func ValidateSpec(spec *v1.Spec) error {
	if spec == nil {
		return ErrEmptySpec
	}
	if err := validateResources(spec.Resources); err != nil {
		return err
	}
	return nil
}

func validateState(state *v1.State) error {
	if state == nil {
		return ErrEmptyState
	}
	if err := validateResources(state.Resources); err != nil {
		return err
	}
	return nil
}

func validateResources(resources v1.Resources) error {
	resourceKeyMap := make(map[string]bool)
	for _, resource := range resources {
		key := resource.ResourceKey()
		if _, ok := resourceKeyMap[key]; ok {
			return fmt.Errorf("%w: %s", ErrDuplicateResourceKey, key)
		}
		resourceKeyMap[key] = true
	}
	return nil
}

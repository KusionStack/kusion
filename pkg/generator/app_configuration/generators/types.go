package generators

import "kusionstack.io/kusion/pkg/models"

type Generator interface {
	Generate(spec *models.Spec) error
}

type NewGeneratorFunc func() (Generator, error)

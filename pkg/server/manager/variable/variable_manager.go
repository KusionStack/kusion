package variable

import (
	"context"
	"errors"
	"strings"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
)

func (v *VariableManager) CreateVariable(ctx context.Context,
	requestPayload request.CreateVariableSetRequest,
) (*entity.Variable, error) {
	// Convert request payload to the domain model.
	var createdEntity entity.Variable
	if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
		return nil, err
	}

	if createdEntity.VariableKey == "" {
		return nil, ErrEmptyVariableKey
	}

	// Compute the fqn for the created entity.
	variableLabelsEntity, err := v.variableLabelsRepo.GetByKey(ctx, createdEntity.VariableKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingVariable
		}

		return nil, err
	}
	createdEntity.Fqn = computeFqn(&createdEntity, variableLabelsEntity)

	// Create variable with repository.
	if err := v.variableRepo.Create(ctx, &createdEntity); err != nil {
		return nil, err
	}

	return &createdEntity, nil
}

func (v *VariableManager) DeleteVariableByFqn(ctx context.Context, fqn string) error {
	if err := v.variableRepo.Delete(ctx, fqn); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingVariable
		}
		return err
	}

	return nil
}

func (v *VariableManager) UpdateVariable(ctx context.Context,
	fqn string, requestPayload request.UpdateVariableSetRequest,
) (*entity.Variable, error) {
	// Convert request payload to domain model.
	var requestEntity entity.Variable
	if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
		return nil, err
	}

	if fqn == "" || requestEntity.Fqn == "" {
		return nil, ErrEmptyVariableFqn
	}

	// Get the existing variable by fqn.
	updatedEntity, err := v.variableRepo.GetByFqn(ctx, fqn)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUpdatingNonExistingVariable
		}

		return nil, err
	}

	// Overwrite non-zero values in request entity to existing entity.
	copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{
		IgnoreEmpty: true,
	})

	// Update variable with repository.
	if err = v.variableRepo.Update(ctx, updatedEntity); err != nil {
		return nil, err
	}

	return updatedEntity, nil
}

func (v *VariableManager) GetVariableByFqn(ctx context.Context, fqn string) (*entity.Variable, error) {
	existingEntity, err := v.variableRepo.GetByFqn(ctx, fqn)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingVariable
		}

		return nil, err
	}

	return existingEntity, nil
}

func (v *VariableManager) ListVariable(ctx context.Context, filter *entity.VariableFilter) (*entity.VariableListResult, error) {
	variableListResult, err := v.variableRepo.List(ctx, filter)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingVariable
		}

		return nil, err
	}

	return variableListResult, nil
}

// computeFqn returns the fqn for a variable.
// NOTE: the calculation rule for fqn is that the lowercase `Key` concatenated with
// the `Labels` in order of priority from low to high.
func computeFqn(variableEntity *entity.Variable, variableLabelsEntity *entity.VariableLabels) string {
	fqnStrList := make([]string, len(variableLabelsEntity.Labels)+1)
	fqnStrList[0] = variableEntity.VariableKey

	for i, labelKey := range variableLabelsEntity.Labels {
		fqnStrList[i+1] = variableEntity.Labels[labelKey]
	}

	return strings.Join(fqnStrList, ":")
}

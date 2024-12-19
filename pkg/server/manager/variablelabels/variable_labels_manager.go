package variablelabels

import (
	"context"
	"errors"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
)

func (vl *VariableLabelsManager) CreateVariableLabels(ctx context.Context,
	requestPayload request.CreateVariableLabelsRequest,
) (*entity.VariableLabels, error) {
	// Convert request payload to the domain model.
	var createdEntity entity.VariableLabels
	if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
		return nil, err
	}

	if createdEntity.VariableKey == "" {
		return nil, ErrEmptyVariableKey
	}
	if len(createdEntity.Labels) == 0 {
		return nil, ErrEmptyVariableLabels
	}

	// Create variable labels with repository.
	if err := vl.variableLabelsRepo.Create(ctx, &createdEntity); err != nil {
		return nil, err
	}

	return &createdEntity, nil
}

func (vl *VariableLabelsManager) DeleteVariableLabelsByKey(ctx context.Context, key string) error {
	if err := vl.variableLabelsRepo.Delete(ctx, key); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingVariable
		}
		return err
	}

	return nil
}

func (vl *VariableLabelsManager) UpdateVariableLabels(ctx context.Context,
	key string, requestPayload request.UpdateVariableLabelsRequest,
) (*entity.VariableLabels, error) {
	// Convert request payload to domain model.
	var requestEntity entity.VariableLabels
	if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
		return nil, err
	}

	if key == "" || requestEntity.VariableKey == "" {
		return nil, ErrEmptyVariableKey
	}
	if len(requestEntity.Labels) == 0 {
		return nil, ErrEmptyVariableLabels
	}

	// Get the existing variable labels by key.
	updatedEntity, err := vl.variableLabelsRepo.GetByKey(ctx, key)
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

	// Update variable labels with repository.
	if err = vl.variableLabelsRepo.Update(ctx, updatedEntity); err != nil {
		return nil, err
	}

	return updatedEntity, nil
}

func (vl *VariableLabelsManager) GetVariableLabelsByKey(ctx context.Context,
	key string,
) (*entity.VariableLabels, error) {
	existingEntity, err := vl.variableLabelsRepo.GetByKey(ctx, key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingVariable
		}

		return nil, err
	}

	return existingEntity, nil
}

func (vl *VariableLabelsManager) ListVariableLabels(ctx context.Context,
	filter *entity.VariableLabelsFilter,
) (*entity.VariableLabelsListResult, error) {
	variableLabelsListResult, err := vl.variableLabelsRepo.List(ctx, filter)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingVariable
		}

		return nil, err
	}

	return variableLabelsListResult, nil
}

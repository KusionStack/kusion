package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/render"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/repository"
	sourcemanager "kusionstack.io/kusion/pkg/server/manager/source"
)

func HandleResult(w http.ResponseWriter, r *http.Request, ctx context.Context, err error, data any) {
	if err != nil {
		render.Render(w, r, FailureResponse(ctx, err))
		return
	}
	render.JSON(w, r, SuccessResponse(ctx, data))
}

func GetSourceByID(ctx context.Context, sourceRepo repository.SourceRepository, id uint) (*entity.Source, error) {
	// Get source by id
	sourceEntity, err := sourceRepo.Get(ctx, id)
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, sourcemanager.ErrGettingNonExistingSource
	} else if err != nil {
		return nil, err
	}
	return sourceEntity, nil
}

func GetProjectByID(ctx context.Context, projectRepo repository.ProjectRepository, id uint) (*entity.Project, error) {
	// Get project by id
	projectEntity, err := projectRepo.Get(ctx, id)
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, ErrProjectDoesNotExist
	} else if err != nil {
		return nil, err
	}
	return projectEntity, nil
}

func GetOrganizationByID(ctx context.Context, organizationRepo repository.OrganizationRepository, id uint) (*entity.Organization, error) {
	// Get organization by id
	organizationEntity, err := organizationRepo.Get(ctx, id)
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, ErrOrganizationDoesNotExist
	} else if err != nil {
		return nil, err
	}
	return organizationEntity, nil
}

func GetStackByID(ctx context.Context, stackRepo repository.StackRepository, id uint) (*entity.Stack, error) {
	// Get stack by id
	stackEntity, err := stackRepo.Get(ctx, id)
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, ErrStackDoesNotExist
	} else if err != nil {
		return nil, err
	}
	return stackEntity, nil
}

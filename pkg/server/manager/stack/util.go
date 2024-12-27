package stack

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
	engineapi "kusionstack.io/kusion/pkg/engine/api"
	sourceapi "kusionstack.io/kusion/pkg/engine/api/source"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform/tfops"
	workspacemanager "kusionstack.io/kusion/pkg/server/manager/workspace"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
	"kusionstack.io/kusion/pkg/util/diff"
)

func BuildOptions(dryrun bool, maxConcurrent int) *engineapi.APIOptions {
	executeOptions := &engineapi.APIOptions{
		// Operator:     "operator",
		// Cluster:      "cluster",
		// IgnoreFields: []string{},
		DryRun:        dryrun,
		MaxConcurrent: maxConcurrent,
		// Watch:         false,
		WatchTimeout: 120,
	}
	return executeOptions
}

// getWorkDirFromSource returns the workdir based on the source
// if the source type is local, it will return the path as an absolute path on the local filesystem
// if the source type is remote (git for example), it will pull the source and return the path to the pulled source
func GetWorkDirFromSource(ctx context.Context, stack *entity.Stack, project *entity.Project) (string, string, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Getting workdir from stack source...")
	// TODO: Also copy the local workdir to /tmp directory?
	var err error
	directory := ""
	workDir := stack.Path

	if project.Source != nil && project.Source.SourceProvider != constant.SourceProviderTypeLocal {
		logger.Info("Non-local source provider, locating pulled source directory")
		// pull the latest source code
		directory, err = sourceapi.Pull(ctx, project.Source)
		if err != nil {
			return "", "", err
		}
		logger.Info("config pulled from source successfully", "directory", directory)
		workDir = filepath.Join(directory, stack.Path)
	}
	return directory, workDir, nil
}

// GetWorkdirAndDirectory is a helper function to get the workdir and directory for a stack
func (m *StackManager) GetWorkdirAndDirectory(ctx context.Context, params *StackRequestParams, stackEntity *entity.Stack) (directory string, workDir string, err error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Getting workdir and directory...")
	if params.ExecuteParams.NoCache {
		// If noCache is set, checkout workdir
		logger.Info("Stack not found in cache. Pulling repo and set cache...")
		directory, workDir, err = GetWorkDirFromSource(ctx, stackEntity, stackEntity.Project)
		if err != nil {
			return "", "", err
		}
		sc := &StackCache{
			LocalDirOnDisk: directory,
			StackPath:      workDir,
		}
		m.repoCache.Set(stackEntity.ID, sc)
	} else {
		// If repoCacheEnv is set, use the cached directory. This takes precedence over the in-memory cache
		repoCacheEnv := os.Getenv("KUSION_SERVER_REPO_CACHE")
		if repoCacheEnv != "" {
			logger.Info("Repo cache found in env var. Using cached directory...")
			directory = repoCacheEnv
			workDir = fmt.Sprintf("%s/%s", directory, stackEntity.Path)
		} else {
			// No env var found, check if stack is in cache
			logger.Info("No repo cache found in env var. Checking cache...")
			stackCache, cacheExists := m.repoCache.Get(stackEntity.ID)
			if cacheExists {
				// if found in repoCache, use the cached workDir and directory
				logger.Info("Stack found in cache. Using cache...")
				workDir = stackCache.StackPath
				directory = stackCache.LocalDirOnDisk
			}
		}
	}
	return directory, workDir, nil
}

func ProcessChanges(ctx context.Context, w http.ResponseWriter, changes *models.Changes, format string, detail bool) (any, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Starting previewing stack in StackManager ...")

	// Mask sensitive data before printing the preview changes.
	for _, v := range changes.ChangeSteps {
		maskedFrom, maskedTo := diff.MaskSensitiveData(v.From, v.To)
		v.From = maskedFrom
		v.To = maskedTo
	}

	if changes.AllUnChange() {
		logger.Info(NoDiffFound)
		return changes, nil
	}

	if format == engineapi.JSONOutput {
		return changes, nil
	}

	// Summary preview table
	changes.Summary(w, true)
	// detail detection
	if detail {
		return changes.Diffs(true), nil
	}
	return "", nil
}

func (m *StackManager) getBackendFromWorkspaceName(ctx context.Context, workspaceName string) (backend.Backend, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Getting backend based on workspace name...")

	var remoteBackend backend.Backend
	if workspaceName == constant.DefaultWorkspace {
		// Get default backend
		if m.defaultBackend.BackendConfig.Type == "" {
			return nil, constant.ErrDefaultBackendNotSet
		}
		return m.getDefaultBackend()
	} else {
		// Get backend by id
		workspaceEntity, err := m.workspaceRepo.GetByName(ctx, workspaceName)
		if err != nil && err == gorm.ErrRecordNotFound {
			return nil, err
		} else if err != nil {
			return nil, err
		}
		// Generate backend from entity
		remoteBackend, err = workspacemanager.NewBackendFromEntity(*workspaceEntity.Backend)
		if err != nil {
			return nil, err
		}
	}
	return remoteBackend, nil
}

func (m *StackManager) metaHelper(
	ctx context.Context,
	id uint,
	workspaceName string,
) (*engineapi.APIOptions, backend.Backend, *v1.Project, *v1.Stack, *v1.Workspace, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Starting getting metadata of the stack in StackManager ...")

	// Get the stack entity by id
	stackEntity, err := m.stackRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, nil, nil, ErrGettingNonExistingStack
		}
		return nil, nil, nil, nil, nil, err
	}

	// Get project by id
	project := stackEntity.Project.ConvertToCore()

	// Get stack by id
	stack := stackEntity.ConvertToCore()

	// Get backend from workspace name
	stackBackend, err := m.getBackendFromWorkspaceName(ctx, workspaceName)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	// Get workspace configurations from backend
	// TODO: should be replaced by variable sets
	wsStorage, err := stackBackend.WorkspaceStorage()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	ws, err := wsStorage.Get(workspaceName)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	executeOptions := BuildOptions(false, constant.MaxConcurrent)

	return executeOptions, stackBackend, project, stack, ws, err
}

func (m *StackManager) getStackProjectAndBackend(ctx context.Context, stackEntity *entity.Stack, workspaceName string) (*v1.Project, *v1.Stack, backend.Backend, error) {
	// Get project by id
	project := stackEntity.Project.ConvertToCore()
	// Get stack by id
	stack := stackEntity.ConvertToCore()
	// Get backend from workspace name
	stateBackend, err := m.getBackendFromWorkspaceName(ctx, workspaceName)
	if err != nil {
		return nil, nil, nil, err
	}
	return project, stack, stateBackend, err
}

func (m *StackManager) getDefaultBackend() (backend.Backend, error) {
	defaultBackendEntity := m.defaultBackend
	remoteBackend, err := workspacemanager.NewBackendFromEntity(defaultBackendEntity)
	if err != nil {
		return nil, err
	}
	return remoteBackend, nil
}

func buildValidStackPath(requestPayload request.CreateStackRequest, projectEntity *entity.Project) (string, bool) {
	stackPath := fmt.Sprintf("%s/%s", projectEntity.Path, requestPayload.Name)
	return stackPath, validStackPath(stackPath)
}

func validStackPath(path string) bool {
	// Validate stack path contains one or more capturing group
	// that contains a backslash with alphanumeric and underscore characters
	return regexp.MustCompile(`^([\/a-zA-Z0-9_-])+$`).MatchString(path)
}

func tempPath(path string) string {
	return fmt.Sprintf("%s/%s", constant.TmpDirPrefix, path)
}

func (m *StackManager) BuildStackFilter(ctx context.Context, query *url.Values) (*entity.StackFilter, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Building stack filter...")

	filter := entity.StackFilter{}

	orgIDParam := query.Get("orgID")
	projectIDParam := query.Get("projectID")
	projectNameParam := query.Get("projectName")
	envParam := query.Get("env")

	if orgIDParam != "" {
		orgID, err := strconv.Atoi(orgIDParam)
		if err != nil {
			return nil, constant.ErrInvalidOrganizationID
		}
		filter.OrgID = uint(orgID)
	}
	// If project id is present, use project id.
	if projectIDParam != "" {
		projectID, err := strconv.Atoi(projectIDParam)
		if err != nil {
			return nil, constant.ErrInvalidProjectID
		}
		filter.ProjectID = uint(projectID)
	} else if projectNameParam != "" {
		// Otherwise, use project name.
		projectEntity, err := m.projectRepo.GetByName(ctx, projectNameParam)
		if err != nil {
			return nil, err
		}
		filter.ProjectID = projectEntity.ID
		if envParam != "" {
			filter.Path = fmt.Sprintf("%s/%s", projectEntity.Path, envParam)
			logger.Info("Showing path filter without cloud", "filter.Path: ", filter.Path)
		}
	}

	// Set pagination parameters.
	page, _ := strconv.Atoi(query.Get("page"))
	if page <= 0 {
		page = constant.CommonPageDefault
	}
	pageSize, _ := strconv.Atoi(query.Get("pageSize"))
	if pageSize <= 0 {
		pageSize = constant.CommonPageSizeDefault
	}
	filter.Pagination = &entity.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return &filter, nil
}

func (m *StackManager) BuildRunFilter(ctx context.Context, query *url.Values) (*entity.RunFilter, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Building run filter...")

	projectIDParam := query.Get("projectID")
	stackIDParam := query.Get("stackID")
	workspaceParam := query.Get("workspace")
	runTypeParam := query.Get("type")
	runStatusParam := query.Get("status")
	startTimeParam := query.Get("startTime")
	endTimeParam := query.Get("endTime")

	filter := entity.RunFilter{}
	if projectIDParam != "" {
		// if project id is present, use project id
		projectID, err := strconv.Atoi(projectIDParam)
		if err != nil {
			return nil, constant.ErrInvalidProjectID
		}
		filter.ProjectID = uint(projectID)
	}
	if stackIDParam != "" {
		// if project id is present, use project id
		stackID, err := strconv.Atoi(stackIDParam)
		if err != nil {
			return nil, constant.ErrInvalidStackID
		}
		filter.StackID = uint(stackID)
	}
	if workspaceParam != "" {
		// if workspace is present, use workspace
		filter.Workspace = workspaceParam
	}
	if runTypeParam != "" {
		// if run type is present, use run type
		filter.Type = strings.Split(runTypeParam, ",")
	}
	if runStatusParam != "" {
		// if run status is present, use run status
		filter.Status = strings.Split(runStatusParam, ",")
	}
	// time format: RFC3339
	if startTimeParam != "" {
		// if start time is present, use start time
		startTime, err := time.Parse(time.RFC3339, startTimeParam)
		if err != nil {
			return nil, err
		}
		filter.StartTime = startTime
	}
	if endTimeParam != "" {
		// if end time is present, use end time
		endTime, err := time.Parse(time.RFC3339, endTimeParam)
		if err != nil {
			return nil, err
		}
		// validate end time is after start time
		if !filter.StartTime.IsZero() && endTime.Before(filter.StartTime) {
			return nil, fmt.Errorf("end time must be after start time")
		}
		filter.EndTime = endTime
	}
	// Set pagination parameters
	page, _ := strconv.Atoi(query.Get("page"))
	if page <= 0 {
		page = constant.CommonPageDefault
	}
	pageSize, _ := strconv.Atoi(query.Get("pageSize"))
	if pageSize <= 0 {
		pageSize = constant.CommonPageSizeDefault
	}
	filter.Pagination = &entity.Pagination{
		Page:     page,
		PageSize: pageSize,
	}
	return &filter, nil
}

func (m *StackManager) ImportTerraformResourceID(ctx context.Context, sp *v1.Spec, importedResources map[string]string) {
	for k, res := range sp.Resources {
		// only for terraform resources
		if res.Type == v1.Terraform {
			// only if kusion resource id exists in the imported resources map
			if cloudResourceID, ok := importedResources[res.ID]; ok {
				res.Extensions[tfops.ImportIDKey] = cloudResourceID
				sp.Resources[k] = res
			}
		}
	}
}

func convertV1ResourceToEntity(resource *v1.Resource) (*entity.Resource, error) {
	// ApiVersion:Kind:Namespace:Name is an idiomatic way for Kubernetes resources.
	// providerNamespace:providerName:resourceType:resourceName for Terraform resources

	// Meta determines whether this is a Kubernetes resource or Terraform resource
	resourceTypeMeta := resource.Type
	var resourceType, resourcePlane, resourceName, resourceProvider string
	var cloudResourceID, iamResourceID, kusionResourceID string
	kusionResourceID = resource.ID

	// Split the resource name to get the parts
	idParts := strings.Split(resource.ID, ":")
	if len(idParts) != 4 {
		// This indicates a Kubernetes resource without the namespace
		if len(idParts) == 3 && isKubernetesResource(resource) {
			modifiedID := fmt.Sprintf("%s:%s:%s:%s", idParts[0], idParts[1], "", idParts[2])
			idParts = strings.Split(modifiedID, ":")
		} else {
			return nil, fmt.Errorf("invalid resource ID: %s", resource.ID)
		}
	}

	// Determine resource plane and resource type based on meta type
	switch resourceTypeMeta {
	case v1.Kubernetes:
		resourcePlane = string(v1.Kubernetes)
		// if this is Kubernetes resource, resource type is apiVersion/kind, resource name is namespace/name
		resourceType = fmt.Sprintf("%s/%s", idParts[0], idParts[1])
		if idParts[2] == "" {
			resourceName = idParts[3]
		} else {
			resourceName = fmt.Sprintf("%s/%s", idParts[2], idParts[3])
		}
	case v1.Terraform:
		// Get provider info for terraform resources
		if providerInfo, ok := resource.Extensions["provider"].(string); ok {
			resourceProvider = providerInfo
		}
		// Look at second element of the id to determine the resource plane
		switch idParts[1] {
		case constant.AWSProviderType:
			resourcePlane = constant.AWSProviderType
			resourceType = idParts[2]
			resourceName = idParts[3]
			if arn, ok := resource.Attributes["arn"].(string); ok {
				cloudResourceID = arn
			}
		case constant.AzureProviderType:
			resourcePlane = constant.AzureProviderType
			resourceType = idParts[2]
			resourceName = idParts[3]
			if resID, ok := resource.Attributes["id"].(string); ok {
				cloudResourceID = resID
			}
		case constant.GoogleProviderType:
			resourcePlane = constant.GoogleProviderType
			resourceType = idParts[2]
			resourceName = idParts[3]
			if resID, ok := resource.Attributes["id"].(string); ok {
				cloudResourceID = resID
			}
		case constant.AliCloudProviderType:
			resourcePlane = constant.AliCloudProviderType
			resourceType = idParts[2]
			resourceName = idParts[3]
			if resID, ok := resource.Attributes["id"].(string); ok {
				cloudResourceID = resID
			}
		default:
			if _, ok := resource.Extensions["provider"]; ok {
				resourcePlane = constant.CustomProviderType
			}
		}
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceTypeMeta)
	}

	return &entity.Resource{
		KusionResourceID: kusionResourceID,
		IAMResourceID:    iamResourceID,
		CloudResourceID:  cloudResourceID,
		ResourcePlane:    resourcePlane,
		ResourceType:     resourceType,
		ResourceName:     resourceName,
		Provider:         resourceProvider,
	}, nil
}

func resourceURN(project, stack, workspace, id string) string {
	return fmt.Sprintf("%s:%s:%s:%s", project, stack, workspace, id)
}

func isKubernetesResource(resource *v1.Resource) bool {
	return resource.Type == v1.Kubernetes
}

func getReleasePath(namespace, source, projectPath, workspace string) string {
	return fmt.Sprintf("%s/%s/%s/%s", namespace, source, projectPath, workspace)
}

func isInRelease(release *v1.Release, id string) bool {
	for _, resource := range release.State.Resources {
		if resource.ID == id {
			return true
		}
	}
	return false
}

func unlockRelease(ctx context.Context, storage release.Storage) error {
	logger := logutil.GetLogger(ctx)
	logger.Info("Getting workdir from stack source...")
	// Get the latest release.
	r, err := release.GetLatestRelease(storage)
	if err != nil {
		return err
	}
	if r == nil {
		logger.Info("No release file found for given stack")
		return nil
	}

	// Update the phase to 'failed', if it was not succeeded or failed.
	if r.Phase != v1.ReleasePhaseSucceeded && r.Phase != v1.ReleasePhaseFailed {
		r.Phase = v1.ReleasePhaseFailed
		if err := storage.Update(r); err != nil {
			return err
		}
		logger.Info("Successfully update release phase!")
		return nil
	} else {
		logger.Info("No need to update the release phase, current phase: ", "phase", r.Phase)
	}
	return nil
}

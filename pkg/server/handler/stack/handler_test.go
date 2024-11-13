package stack

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/infra/persistence"
	"kusionstack.io/kusion/pkg/server/handler"
	stackmanager "kusionstack.io/kusion/pkg/server/manager/stack"
)

func TestStackHandler(t *testing.T) {
	var (
		stackName        = "test-stack"
		stackNameSecond  = "test-stack-2"
		projectName      = "test-project"
		projectPath      = "/path/to/project"
		stackPath        = "/path/to/stack"
		stackNameUpdated = "test-stack-updated"
		stackPathUpdated = "/path/to/stacks/updated"
		owners           = persistence.MultiString{"hua.li", "xiaoming.li"}
	)
	t.Run("ListStacks", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, stackHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "sync_state", "Project__id", "Project__name", "Project__path"}).
				AddRow(1, stackName, stackPath, constant.StackStateUnSynced, 1, projectName, projectPath).
				AddRow(2, stackNameSecond, stackPath, constant.StackStateUnSynced, 2, projectName, projectPath))

		// Create a new HTTP request
		req, err := http.NewRequest("GET", "/stacks", nil)
		assert.NoError(t, err)

		// Call the ListStacks handler function
		stackHandler.ListStacks()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshall the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Assertion
		assert.Equal(t, 2, len(resp.Data.([]any)))
	})

	t.Run("GetStack", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, stackHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "sync_state", "Project__id", "Project__name", "Project__path"}).
				AddRow(1, stackName, stackPath, constant.StackStateUnSynced, 1, projectName, projectPath))

		// Create a new HTTP request
		req, err := http.NewRequest("GET", "/stacks/{stackID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("stackID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Call the ListStacks handler function
		stackHandler.GetStack()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshal the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Assertion
		assert.Equal(t, float64(1), resp.Data.(map[string]any)["id"])
		assert.Equal(t, stackName, resp.Data.(map[string]any)["name"])
		assert.Equal(t, stackPath, resp.Data.(map[string]any)["path"])
		assert.Equal(t, float64(1), resp.Data.(map[string]any)["project"].(map[string]any)["id"])
	})

	t.Run("CreateStack", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, stackHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Create a new HTTP request
		req, err := http.NewRequest("POST", "/stacks", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Set request body
		requestPayload := request.CreateStackRequest{
			Name:           stackName,
			Path:           stackPath,
			DesiredVersion: "latest",
			ProjectID:      1,
		}
		reqBody, err := json.Marshal(requestPayload)
		assert.NoError(t, err)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
		req.Header.Add("Content-Type", "application/json")

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "Organization__id", "Organization__name", "Organization__owners", "Source__id", "Source__name", "Source__remote", "Source__source_provider"}).
				AddRow(1, projectName, projectPath, 1, "test-org", owners, 1, "test-source", "https://github.com/test/repo", constant.SourceProviderTypeGithub))
		sqlMock.ExpectBegin()
		sqlMock.ExpectExec("INSERT").
			WillReturnResult(sqlmock.NewResult(int64(1), int64(1)))
		sqlMock.ExpectCommit()

		// Call the CreateStack handler function
		stackHandler.CreateStack()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshal the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Assertion
		assert.Equal(t, float64(1), resp.Data.(map[string]any)["id"])
		assert.Equal(t, stackName, resp.Data.(map[string]any)["name"])
		assert.Equal(t, stackPath, resp.Data.(map[string]any)["path"])
		assert.Equal(t, "latest", resp.Data.(map[string]any)["desiredVersion"])
		assert.Equal(t, float64(1), resp.Data.(map[string]any)["project"].(map[string]any)["id"])
	})

	t.Run("UpdateExistingStack", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, stackHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Update a new HTTP request
		req, err := http.NewRequest("PUT", "/stacks/{stackID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("stackID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Set request body
		requestPayload := request.UpdateStackRequest{
			// Set your request payload fields here
			ID: 1,
			CreateStackRequest: request.CreateStackRequest{
				Name:      stackNameUpdated,
				Path:      stackPathUpdated,
				ProjectID: 1,
			},
		}
		reqBody, err := json.Marshal(requestPayload)
		assert.NoError(t, err)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
		req.Header.Add("Content-Type", "application/json")

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "Organization__id", "Organization__name", "Organization__owners", "Source__id", "Source__remote", "Source__source_provider"}).
				AddRow(1, stackName, stackPath, 1, "test-org", owners, 1, "https://github.com/test/repo", constant.SourceProviderTypeGithub))
		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "sync_state", "Project__id", "Project__name", "Project__path"}).
				AddRow(1, stackName, stackPath, constant.StackStateUnSynced, 1, projectName, projectPath))
		sqlMock.ExpectExec("UPDATE").
			WillReturnResult(sqlmock.NewResult(int64(1), int64(1)))

		// Call the ListStacks handler function
		stackHandler.UpdateStack()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshall the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Assertion
		assert.Equal(t, float64(1), resp.Data.(map[string]any)["id"])
		assert.Equal(t, stackNameUpdated, resp.Data.(map[string]any)["name"])
		assert.Equal(t, stackPathUpdated, resp.Data.(map[string]any)["path"])
	})

	t.Run("Delete Existing Stack", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, stackHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Create a new HTTP request
		req, err := http.NewRequest("DELETE", "/stacks/{stackID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("stackID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Mock the Delete method of the stack repository
		sqlMock.ExpectBegin()
		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				AddRow(1))
		sqlMock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 0))
		sqlMock.ExpectCommit()

		// Call the DeleteStack handler function
		stackHandler.DeleteStack()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshall the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.Equal(t, "Deletion Success", resp.Data)
	})

	t.Run("Delete Nonexisting Stack", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, stackHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Create a new HTTP request
		req, err := http.NewRequest("DELETE", "/stacks/{stackID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("stackID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		sqlMock.ExpectBegin()
		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// Call the DeleteStack handler function
		stackHandler.DeleteStack()(recorder, req)
		// Unmarshall the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, false, resp.Success)
		assert.Equal(t, stackmanager.ErrGettingNonExistingStack.Error(), resp.Message)
	})

	t.Run("Update Nonexisting Stack", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, stackHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Update a new HTTP request
		req, err := http.NewRequest("POST", "/stacks/{stackID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("stackID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Set request body
		requestPayload := request.UpdateStackRequest{
			// Set your request payload fields here
			ID: 1,
			CreateStackRequest: request.CreateStackRequest{
				Name: "test-stack-updated",
				Path: stackPathUpdated,
			},
		}
		reqBody, err := json.Marshal(requestPayload)
		assert.NoError(t, err)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
		req.Header.Add("Content-Type", "application/json")

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "Organization__id", "Organization__name", "Organization__owners", "Source__id", "Source__remote", "Source__source_provider"}).
				AddRow(1, stackName, stackPath, 1, "test-org", owners, 1, "https://github.com/test/repo", constant.SourceProviderTypeGithub))
		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// Call the UpdateStack handler function
		stackHandler.UpdateStack()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshall the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Assertion
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, false, resp.Success)
		assert.Equal(t, stackmanager.ErrUpdatingNonExistingStack.Error(), resp.Message)
	})
}

func setupTest(t *testing.T) (sqlmock.Sqlmock, *gorm.DB, *httptest.ResponseRecorder, *Handler) {
	fakeGDB, sqlMock, err := persistence.GetMockDB()
	require.NoError(t, err)
	stackRepo := persistence.NewStackRepository(fakeGDB)
	projectRepo := persistence.NewProjectRepository(fakeGDB)
	workspaceRepo := persistence.NewWorkspaceRepository(fakeGDB)
	resourceRepo := persistence.NewResourceRepository(fakeGDB)
	runRepo := persistence.NewRunRepository(fakeGDB)
	stackHandler := &Handler{
		stackManager: stackmanager.NewStackManager(stackRepo, projectRepo, workspaceRepo, resourceRepo, runRepo, entity.Backend{}, constant.MaxConcurrent),
	}
	recorder := httptest.NewRecorder()
	return sqlMock, fakeGDB, recorder, stackHandler
}

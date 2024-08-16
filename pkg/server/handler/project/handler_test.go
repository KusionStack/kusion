package project

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	projectmanager "kusionstack.io/kusion/pkg/server/manager/project"
)

func TestProjectHandler(t *testing.T) {
	var (
		projectName        = "test-project"
		projectNameSecond  = "test-project-2"
		projectPath        = "/path/to/project"
		projectNameUpdated = "test-project-updated"
		projectPathUpdated = "/path/to/projects/updated"
		owners             = persistence.MultiString{"hua.li", "xiaoming.li"}
	)
	t.Run("ListProjects", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, projectHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "Organization__id", "Organization__name", "Organization__owners", "Source__id", "Source__remote", "Source__source_provider"}).
				AddRow(1, projectName, projectPath, 1, "test-org", owners, 1, "https://github.com/test/repo", constant.SourceProviderTypeGithub).
				AddRow(2, projectNameSecond, projectPath, 2, "test-org-2", owners, 1, "https://github.com/test/repo", constant.SourceProviderTypeGithub))

		// Create a new HTTP request
		req, err := http.NewRequest("GET", "/projects", nil)
		assert.NoError(t, err)

		// Call the ListProjects handler function
		projectHandler.ListProjects()(recorder, req)
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

	t.Run("GetProject", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, projectHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "Organization__id", "Organization__name", "Organization__owners", "Source__id", "Source__remote", "Source__source_provider"}).
				AddRow(1, projectName, projectPath, 1, "test-org", owners, 1, "https://github.com/test/repo", constant.SourceProviderTypeGithub))

		// Create a new HTTP request
		req, err := http.NewRequest("GET", "/projects/{projectID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("projectID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Call the ListProjects handler function
		projectHandler.GetProject()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshal the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Assertion
		assert.Equal(t, float64(1), resp.Data.(map[string]any)["id"])
		assert.Equal(t, projectName, resp.Data.(map[string]any)["name"])
	})

	t.Run("CreateProject", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, projectHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Create a new HTTP request
		req, err := http.NewRequest("POST", "/projects", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Set request body
		requestPayload := request.CreateProjectRequest{
			Name:           projectName,
			Path:           projectPath,
			SourceID:       uint(1),
			OrganizationID: uint(1),
		}
		reqBody, err := json.Marshal(requestPayload)
		assert.NoError(t, err)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
		req.Header.Add("Content-Type", "application/json")

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "remote", "source_provider"}).
				AddRow(1, "https://github.com/test/repo", constant.SourceProviderTypeGithub))
		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owners"}).
				AddRow(1, "test-org", owners))
		sqlMock.ExpectBegin()
		sqlMock.ExpectExec("INSERT").
			WillReturnResult(sqlmock.NewResult(int64(1), int64(1)))
		sqlMock.ExpectCommit()

		// Call the CreateProject handler function
		projectHandler.CreateProject()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshal the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Assertion
		assert.Equal(t, float64(1), resp.Data.(map[string]any)["id"])
		assert.Equal(t, projectName, resp.Data.(map[string]any)["name"])
	})

	t.Run("UpdateExistingProject", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, projectHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Update a new HTTP request
		req, err := http.NewRequest("PUT", "/projects/{projectID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("projectID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Set request body
		requestPayload := request.UpdateProjectRequest{
			// Set your request payload fields here
			ID: 1,
			CreateProjectRequest: request.CreateProjectRequest{
				Name:           projectNameUpdated,
				Path:           projectPathUpdated,
				OrganizationID: 1,
			},
		}
		reqBody, err := json.Marshal(requestPayload)
		assert.NoError(t, err)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
		req.Header.Add("Content-Type", "application/json")

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owners"}).
				AddRow(1, "test-org", owners))
		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "Organization__id", "Organization__name", "Organization__owners", "Source__id", "Source__remote", "Source__source_provider"}).
				AddRow(1, projectName, projectPath, 1, "test-org", owners, 1, "https://github.com/test/repo", constant.SourceProviderTypeGithub))
		sqlMock.ExpectExec("UPDATE").
			WillReturnResult(sqlmock.NewResult(int64(1), int64(1)))

		// Call the ListProjects handler function
		projectHandler.UpdateProject()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		fmt.Println(recorder.Body.String())

		// Unmarshall the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Assertion
		assert.Equal(t, float64(1), resp.Data.(map[string]any)["id"])
		assert.Equal(t, projectNameUpdated, resp.Data.(map[string]any)["name"])
		assert.Equal(t, projectPathUpdated, resp.Data.(map[string]any)["path"])
	})

	t.Run("Delete Existing Project", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, projectHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Create a new HTTP request
		req, err := http.NewRequest("DELETE", "/projects/{projectID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("projectID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Mock the Delete method of the project repository
		sqlMock.ExpectBegin()
		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				AddRow(1))
		sqlMock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 0))
		sqlMock.ExpectCommit()

		// Call the DeleteProject handler function
		projectHandler.DeleteProject()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshall the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.Equal(t, "Deletion Success", resp.Data)
	})

	t.Run("Delete Nonexisting Project", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, projectHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Create a new HTTP request
		req, err := http.NewRequest("DELETE", "/projects/{projectID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("projectID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		sqlMock.ExpectBegin()
		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// Call the DeleteProject handler function
		projectHandler.DeleteProject()(recorder, req)
		// Unmarshall the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, false, resp.Success)
		assert.Equal(t, projectmanager.ErrGettingNonExistingProject.Error(), resp.Message)
	})

	t.Run("Update Nonexisting Project", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, projectHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Update a new HTTP request
		req, err := http.NewRequest("POST", "/projects/{projectID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("projectID", "2")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Set request body
		requestPayload := request.UpdateProjectRequest{
			// Set your request payload fields here
			ID: 2,
			CreateProjectRequest: request.CreateProjectRequest{
				Name:           "test-project-updated",
				Path:           projectPathUpdated,
				OrganizationID: 1,
			},
		}
		reqBody, err := json.Marshal(requestPayload)
		assert.NoError(t, err)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
		req.Header.Add("Content-Type", "application/json")

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// Call the UpdateProject handler function
		projectHandler.UpdateProject()(recorder, req)
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
		assert.Equal(t, projectmanager.ErrUpdatingNonExistingProject.Error(), resp.Message)
	})
}

func setupTest(t *testing.T) (sqlmock.Sqlmock, *gorm.DB, *httptest.ResponseRecorder, *Handler) {
	fakeGDB, sqlMock, err := persistence.GetMockDB()
	require.NoError(t, err)
	projectRepo := persistence.NewProjectRepository(fakeGDB)
	sourceRepo := persistence.NewSourceRepository(fakeGDB)
	organizationRepo := persistence.NewOrganizationRepository(fakeGDB)
	projectHandler := &Handler{
		projectManager: projectmanager.NewProjectManager(projectRepo, organizationRepo, sourceRepo, entity.Source{}),
	}
	recorder := httptest.NewRecorder()
	return sqlMock, fakeGDB, recorder, projectHandler
}

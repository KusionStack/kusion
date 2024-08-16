package workspace

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
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/infra/persistence"
	"kusionstack.io/kusion/pkg/server/handler"
	workspacemanager "kusionstack.io/kusion/pkg/server/manager/workspace"
)

func TestWorkspaceHandler(t *testing.T) {
	var (
		wsName        = "test-ws"
		wsNameUpdated = "test-ws-updated"
	)
	t.Run("ListWorkspaces", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, workspaceHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "Backend__id"}).
				AddRow(1, "test-ws", 1).
				AddRow(2, "test-ws-2", 2))

		// Create a new HTTP request
		req, err := http.NewRequest("GET", "/workspaces", nil)
		assert.NoError(t, err)

		// Call the ListWorkspaces handler function
		workspaceHandler.ListWorkspaces()(recorder, req)
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

	t.Run("GetWorkspace", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, workspaceHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "Backend__id"}).
				AddRow(1, wsName, 1))

		// Create a new HTTP request
		req, err := http.NewRequest("GET", "/workspaces/{workspaceID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("workspaceID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Call the ListWorkspaces handler function
		workspaceHandler.GetWorkspace()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshal the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Assertion
		assert.Equal(t, float64(1), resp.Data.(map[string]any)["id"])
		assert.Equal(t, wsName, resp.Data.(map[string]any)["name"])
	})

	t.Run("CreateWorkspace", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, workspaceHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Create a new HTTP request
		req, err := http.NewRequest("POST", "/workspaces", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("workspaceID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Set request body
		requestPayload := request.CreateWorkspaceRequest{
			Name:      wsName,
			BackendID: 1,
			Owners:    []string{"hua.li", "xiaoming.li"},
		}
		reqBody, err := json.Marshal(requestPayload)
		assert.NoError(t, err)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
		req.Header.Add("Content-Type", "application/json")

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				AddRow(1))
		sqlMock.ExpectBegin()
		sqlMock.ExpectExec("INSERT").
			WillReturnResult(sqlmock.NewResult(int64(1), int64(1)))
		sqlMock.ExpectCommit()

		// Call the CreateWorkspace handler function
		workspaceHandler.CreateWorkspace()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshal the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Assertion
		assert.Equal(t, float64(1), resp.Data.(map[string]any)["id"])
		assert.Equal(t, wsName, resp.Data.(map[string]any)["name"])
	})

	t.Run("UpdateExistingWorkspace", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, workspaceHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Update a new HTTP request
		req, err := http.NewRequest("POST", "/workspaces/{workspaceID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("workspaceID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Set request body
		requestPayload := request.UpdateWorkspaceRequest{
			ID:        1,
			Name:      wsNameUpdated,
			BackendID: 1,
		}
		reqBody, err := json.Marshal(requestPayload)
		assert.NoError(t, err)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
		req.Header.Add("Content-Type", "application/json")

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "Backend__id"}).
				AddRow(1, "test-ws-updated", 1))
		sqlMock.ExpectExec("UPDATE").
			WillReturnResult(sqlmock.NewResult(int64(1), int64(1)))

		// Call the ListWorkspaces handler function
		workspaceHandler.UpdateWorkspace()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshall the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Assertion
		assert.Equal(t, float64(1), resp.Data.(map[string]any)["id"])
		assert.Equal(t, wsNameUpdated, resp.Data.(map[string]any)["name"])
	})

	t.Run("Delete Existing Workspace", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, workspaceHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Create a new HTTP request
		req, err := http.NewRequest("DELETE", "/workspaces/{workspaceID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("workspaceID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Mock the Delete method of the workspace repository
		sqlMock.ExpectBegin()
		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				AddRow(1))
		sqlMock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 0))
		sqlMock.ExpectCommit()

		// Call the DeleteWorkspace handler function
		workspaceHandler.DeleteWorkspace()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshall the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.Equal(t, "Deletion Success", resp.Data)
	})

	t.Run("Delete Nonexisting Workspace", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, workspaceHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Create a new HTTP request
		req, err := http.NewRequest("DELETE", "/workspaces/{workspaceID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("workspaceID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		sqlMock.ExpectBegin()
		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// Call the DeleteWorkspace handler function
		workspaceHandler.DeleteWorkspace()(recorder, req)
		// Unmarshall the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, false, resp.Success)
		assert.Equal(t, workspacemanager.ErrGettingNonExistingWorkspace.Error(), resp.Message)
	})

	t.Run("Update Nonexisting Workspace", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, workspaceHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Update a new HTTP request
		req, err := http.NewRequest("POST", "/workspaces/{workspaceID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("workspaceID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Set request body
		requestPayload := request.UpdateWorkspaceRequest{
			// Set your request payload fields here
			ID:        1,
			Name:      "test-ws-updated",
			BackendID: 1,
		}
		reqBody, err := json.Marshal(requestPayload)
		assert.NoError(t, err)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
		req.Header.Add("Content-Type", "application/json")

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// Call the UpdateWorkspace handler function
		workspaceHandler.UpdateWorkspace()(recorder, req)
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
		assert.Equal(t, workspacemanager.ErrUpdatingNonExistingWorkspace.Error(), resp.Message)
	})
}

func setupTest(t *testing.T) (sqlmock.Sqlmock, *gorm.DB, *httptest.ResponseRecorder, *Handler) {
	fakeGDB, sqlMock, err := persistence.GetMockDB()
	require.NoError(t, err)
	workspaceRepo := persistence.NewWorkspaceRepository(fakeGDB)
	backendRepo := persistence.NewBackendRepository(fakeGDB)
	workspaceHandler := &Handler{
		workspaceManager: workspacemanager.NewWorkspaceManager(workspaceRepo, backendRepo, entity.Backend{}),
	}
	recorder := httptest.NewRecorder()
	return sqlMock, fakeGDB, recorder, workspaceHandler
}

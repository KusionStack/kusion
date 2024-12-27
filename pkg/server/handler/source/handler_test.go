package source

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
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/infra/persistence"
	"kusionstack.io/kusion/pkg/server/handler"
	sourcemanager "kusionstack.io/kusion/pkg/server/manager/source"
)

func TestSourceHandler(t *testing.T) {
	t.Run("ListSources", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, sourceHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		sqlMock.ExpectQuery("SELECT count(.*) FROM `source`").
			WillReturnRows(
				sqlmock.NewRows([]string{"count"}).
					AddRow(2))

		sqlMock.ExpectQuery("SELECT .* FROM `source`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "source_provider"}).
				AddRow(1, string(constant.SourceProviderTypeGithub)).
				AddRow(2, string(constant.SourceProviderTypeLocal)))

		// Create a new HTTP request
		req, err := http.NewRequest("GET", "/sources", nil)
		assert.NoError(t, err)

		// Call the ListSources handler function
		sourceHandler.ListSources()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshall the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Assertion
		assert.Equal(t, 2, len(resp.Data.(map[string]any)["sources"].([]any)))
		assert.Equal(t, string(constant.SourceProviderTypeGithub), resp.Data.(map[string]any)["sources"].([]any)[0].(map[string]any)["sourceProvider"])
		assert.Equal(t, string(constant.SourceProviderTypeLocal), resp.Data.(map[string]any)["sources"].([]any)[1].(map[string]any)["sourceProvider"])
	})

	t.Run("GetSource", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, sourceHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "source_provider"}).
				AddRow(1, string(constant.SourceProviderTypeGithub)))

		// Create a new HTTP request
		req, err := http.NewRequest("GET", "/sources/{sourceID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("sourceID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Call the ListSources handler function
		sourceHandler.GetSource()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshal the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Assertion
		assert.Equal(t, float64(1), resp.Data.(map[string]any)["id"])
		assert.Equal(t, string(constant.SourceProviderTypeGithub), resp.Data.(map[string]any)["sourceProvider"])
	})

	t.Run("CreateSource", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, sourceHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Create a new HTTP request
		req, err := http.NewRequest("POST", "/sources", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("sourceID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Set request body
		requestPayload := request.CreateSourceRequest{
			// Set your request payload fields here
			Name:           "test-source",
			SourceProvider: string(constant.SourceProviderTypeGithub),
			Remote:         "https://github.com/test/remote",
		}
		reqBody, err := json.Marshal(requestPayload)
		assert.NoError(t, err)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
		req.Header.Add("Content-Type", "application/json")

		sqlMock.ExpectBegin()
		sqlMock.ExpectExec("INSERT").
			WillReturnResult(sqlmock.NewResult(int64(1), int64(1)))
		sqlMock.ExpectCommit()

		// Call the CreateSource handler function
		sourceHandler.CreateSource()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshal the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Assertion
		assert.Equal(t, float64(1), resp.Data.(map[string]any)["id"])
		assert.Equal(t, string(constant.SourceProviderTypeGithub), resp.Data.(map[string]any)["sourceProvider"])
	})

	t.Run("UpdateExistingSource", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, sourceHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Update a new HTTP request
		req, err := http.NewRequest("POST", "/sources/{sourceID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("sourceID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Set request body
		requestPayload := request.UpdateSourceRequest{
			// Set your request payload fields here
			ID: 1,
			CreateSourceRequest: request.CreateSourceRequest{
				SourceProvider: string(constant.SourceProviderTypeGithub),
				Remote:         "https://github.com/test/updated-remote",
			},
		}
		reqBody, err := json.Marshal(requestPayload)
		assert.NoError(t, err)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
		req.Header.Add("Content-Type", "application/json")

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "source_provider"}).
				AddRow(1, constant.SourceProviderTypeGithub))
		sqlMock.ExpectExec("UPDATE").
			WillReturnResult(sqlmock.NewResult(int64(1), int64(1)))

		// Call the ListSources handler function
		sourceHandler.UpdateSource()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshall the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Assertion
		assert.Equal(t, float64(1), resp.Data.(map[string]any)["id"])
		assert.Equal(t, string(constant.SourceProviderTypeGithub), resp.Data.(map[string]any)["sourceProvider"])
		assert.Equal(t, "/test/updated-remote", resp.Data.(map[string]any)["remote"].(map[string]any)["Path"])
	})

	t.Run("Delete Existing Source", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, sourceHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Create a new HTTP request
		req, err := http.NewRequest("DELETE", "/sources/{sourceID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("sourceID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Mock the Delete method of the source repository
		sqlMock.ExpectBegin()
		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				AddRow(1))
		sqlMock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 0))
		sqlMock.ExpectCommit()

		// Call the DeleteSource handler function
		sourceHandler.DeleteSource()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshall the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.Equal(t, "Deletion Success", resp.Data)
	})

	t.Run("Delete Nonexisting Source", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, sourceHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Create a new HTTP request
		req, err := http.NewRequest("DELETE", "/sources/{sourceID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("sourceID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		sqlMock.ExpectBegin()
		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// Call the DeleteSource handler function
		sourceHandler.DeleteSource()(recorder, req)
		// Unmarshall the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, resp.Success, false)
		assert.Equal(t, resp.Message, sourcemanager.ErrGettingNonExistingSource.Error())
	})

	t.Run("Update Nonexisting Source", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, sourceHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		// Update a new HTTP request
		req, err := http.NewRequest("POST", "/sources/{sourceID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("sourceID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Set request body
		requestPayload := request.UpdateSourceRequest{
			// Set your request payload fields here
			ID: 1,
			CreateSourceRequest: request.CreateSourceRequest{
				SourceProvider: string(constant.SourceProviderTypeGithub),
				Remote:         "https://github.com/test/updated-remote",
			},
		}
		reqBody, err := json.Marshal(requestPayload)
		assert.NoError(t, err)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
		req.Header.Add("Content-Type", "application/json")

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// Call the UpdateSource handler function
		sourceHandler.UpdateSource()(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshall the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Assertion
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, resp.Success, false)
		assert.Equal(t, resp.Message, sourcemanager.ErrUpdatingNonExistingSource.Error())
	})
}

func setupTest(t *testing.T) (sqlmock.Sqlmock, *gorm.DB, *httptest.ResponseRecorder, *Handler) {
	fakeGDB, sqlMock, err := persistence.GetMockDB()
	require.NoError(t, err)
	repo := persistence.NewSourceRepository(fakeGDB)
	sourceHandler := &Handler{
		sourceManager: sourcemanager.NewSourceManager(repo),
	}
	recorder := httptest.NewRecorder()
	return sqlMock, fakeGDB, recorder, sourceHandler
}

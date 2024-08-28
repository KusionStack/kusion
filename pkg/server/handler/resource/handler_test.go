package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/infra/persistence"
	"kusionstack.io/kusion/pkg/server/handler"
	resourcemanager "kusionstack.io/kusion/pkg/server/manager/resource"
)

func TestResourceHandler(t *testing.T) {
	var (
		resourceName       = "test-resource"
		resourceNameSecond = "test-resource-2"
	)
	t.Run("ListResources", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, stackHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "resource_type", "resource_plane", "resource_name", "kusion_resource_id"}).
				AddRow(1, "Kubernetes", "Kubernetes", resourceName, "a:b:c:d").
				AddRow(2, "Terraform", "AWS", resourceNameSecond, "e:f:g:h"))

		// Create a new HTTP request
		req, err := http.NewRequest("GET", "/resources", nil)
		assert.NoError(t, err)

		// Call the ListResources handler function
		stackHandler.ListResources()(recorder, req)
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

	t.Run("GetResource", func(t *testing.T) {
		sqlMock, fakeGDB, recorder, resourceHandler := setupTest(t)
		defer persistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "resource_type", "resource_plane", "resource_name", "kusion_resource_id"}).
				AddRow(1, "Kubernetes", "Kubernetes", resourceName, "a:b:c:d"))

		// Create a new HTTP request
		req, err := http.NewRequest("GET", "/resources/{resourceID}", nil)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("resourceID", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Call the ListResources handler function
		resourceHandler.GetResource()(recorder, req)
		fmt.Println(recorder.Body.String())
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Unmarshal the response body
		var resp handler.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Assertion
		assert.Equal(t, float64(1), resp.Data.(map[string]any)["id"])
		assert.Equal(t, resourceName, resp.Data.(map[string]any)["resourceName"])
		assert.Equal(t, "Kubernetes", resp.Data.(map[string]any)["resourceType"])
		assert.Equal(t, "Kubernetes", resp.Data.(map[string]any)["resourcePlane"])
	})
}

func setupTest(t *testing.T) (sqlmock.Sqlmock, *gorm.DB, *httptest.ResponseRecorder, *Handler) {
	fakeGDB, sqlMock, err := persistence.GetMockDB()
	require.NoError(t, err)
	resourceRepo := persistence.NewResourceRepository(fakeGDB)
	resourceHandler := &Handler{
		resourceManager: resourcemanager.NewResourceManager(resourceRepo),
	}
	recorder := httptest.NewRecorder()
	return sqlMock, fakeGDB, recorder, resourceHandler
}

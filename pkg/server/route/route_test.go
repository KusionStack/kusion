package route

import (
	"testing"

	"github.com/go-chi/chi/v5"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/server"
)

// TestNewCoreRoute will test the NewCoreRoute function with different
// configurations.
// func TestNewCoreRoute(t *testing.T) {
// 	// Mock the NewSearchStorage function to return a mock storage instead of
// 	// actual implementation.

// 	fakeGDB, _, err := persistence.GetMockDB()
// 	require.NoError(t, err)
// 	tests := []struct {
// 		name         string
// 		config       server.Config
// 		expectError  bool
// 		expectRoutes []string
// 	}{
// 		{
// 			name: "route test",
// 			config: server.Config{
// 				DB:             fakeGDB,
// 				LogFilePath:    "test.log",
// 				AuthEnabled:    false,
// 				AuthWhitelist:  []string{"kusion"},
// 				AuthKeyType:    "RS256",
// 				AutoMigrate:    false,
// 				DefaultBackend: entity.Backend{},
// 				DefaultSource:  entity.Source{},
// 				MaxConcurrent:  10,
// 			},
// 			expectError: false,
// 			expectRoutes: []string{
// 				"/endpoints",
// 				"/server-configs",
// 				"/api/v1/stack",
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			router, err := NewCoreRoute(&tt.config)
// 			if tt.expectError {
// 				require.Error(t, err)
// 			} else {
// 				require.NoError(t, err)
// 				for _, route := range tt.expectRoutes {
// 					req := httptest.NewRequest(http.MethodGet, route, nil)
// 					rr := httptest.NewRecorder()
// 					router.ServeHTTP(rr, req)

// 					// Assert status code is not 404 to ensure the route exists.
// 					require.Equal(t, http.StatusOK, rr.Code)
// 					require.NotEqual(t, http.StatusNotFound, rr.Code, "Route should exist: %s", route)
// 				}
// 			}
// 		})
// 	}
// }

func TestSetupRestAPIV1(t *testing.T) {
	config := &server.Config{
		AuthEnabled:    false,
		AuthWhitelist:  []string{"example.com"},
		AuthKeyType:    "RS256",
		DB:             nil, // Replace with your mock DB
		AutoMigrate:    false,
		DefaultBackend: entity.Backend{},
		DefaultSource:  entity.Source{},
		MaxConcurrent:  10,
	}

	r := chi.NewRouter()
	setupRestAPIV1(r, config)

	// Add your assertions here
}

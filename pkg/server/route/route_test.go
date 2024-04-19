package route

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/stretchr/testify/require"
// 	"kusionstack.io/kusion/pkg/infra/persistence"
// 	"kusionstack.io/kusion/pkg/server"
// )

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
// 				DB: fakeGDB,
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
// 					request, _ := json.Marshal(req.URL)
// 					fmt.Println(string(request))
// 					rr := httptest.NewRecorder()
// 					router.ServeHTTP(rr, req)
// 					fmt.Println(rr.Code)
// 					fmt.Println(rr.Header())

// 					// Assert status code is not 404 to ensure the route exists.
// 					require.Equal(t, http.StatusOK, rr.Code)
// 					require.NotEqual(t, http.StatusNotFound, rr.Code, "Route should exist: %s", route)
// 				}
// 			}
// 		})
// 	}
// }

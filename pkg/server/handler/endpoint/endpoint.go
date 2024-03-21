package endpoint

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-logr/logr"
	"kusionstack.io/kusion/pkg/server/util"
)

// Endpoints provides an endpoint to list all available endpoints registered
// in the router.
//
// @Summary      List all available endpoints
// @Description  List all registered endpoints in the router
// @Tags         debug
// @Accept       plain
// @Produce      plain
// @Success      200  {string}  string  "Endpoints listed successfully"
// @Router       /endpoints [get]
func Endpoints(router chi.Router) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := util.GetLogger(r.Context())
		endpoints := listEndpoints(log, router)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(strings.Join(endpoints, "\n")))
	}
}

// listEndpoints generates a list of all routes registered in the router.
func listEndpoints(log logr.Logger, r chi.Router) []string {
	var endpoints []string

	// Walk through the routes to collect endpoints
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		endpoint := fmt.Sprintf("%s\t%s", method, route)
		endpoints = append(endpoints, endpoint)
		return nil
	}

	// Populate the list of endpoints by walking through the router
	if err := chi.Walk(r, walkFunc); err != nil {
		log.Error(err, "Walking routes error")
	}

	// Sort the collected endpoints alphabetically
	sort.Strings(endpoints)
	return endpoints
}

package route

import (
	"context"
	"expvar"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpswagger "github.com/swaggo/http-swagger"
	"github.com/swaggo/swag/example/basic/docs"
	endpointhandler "kusionstack.io/kusion/pkg/server/handler/endpoint"
	stackhandler "kusionstack.io/kusion/pkg/server/handler/stack"
	appmiddleware "kusionstack.io/kusion/pkg/server/middleware"

	"kusionstack.io/kusion/pkg/server/util"
)

// NewCoreRoute creates and configures an instance of chi.Mux with the given
// configuration and extra configuration parameters.
func NewCoreRoute() (*chi.Mux, error) {
	router := chi.NewRouter()

	// Set up middlewares for logging, recovery, and timing, etc.
	router.Use(middleware.RequestID)
	router.Use(appmiddleware.DefaultLogger)
	router.Use(appmiddleware.APILogger)
	router.Use(appmiddleware.Timing)
	router.Use(middleware.Recoverer)

	// Set up the API routes for version 1 of the API.
	router.Route("/api/v1", func(r chi.Router) {
		setupRestAPIV1(r)
	})

	// Set up the root routes.
	docs.SwaggerInfo.BasePath = "/"
	router.Get("/docs/*", httpswagger.Handler())

	// Endpoint to list all available endpoints in the router.
	router.Get("/endpoints", endpointhandler.Endpoints(router))

	// Endpoint to list all available endpoints in the router.
	router.Get("/server-configs", expvar.Handler().ServeHTTP)
	http.ListenAndServe("localhost:8888", router)

	logger := util.GetLogger(context.TODO())
	logger.Info("Server Started...")

	return router, nil
}

// setupRestAPIV1 configures routing for the API version 1, grouping routes by
// resource type and setting up proper handlers.
func setupRestAPIV1(
	r chi.Router,
) {
	r.Route("/project", func(r chi.Router) {
		r.Route("/{projectName}", func(r chi.Router) {
			r.Route("/stack", func(r chi.Router) {
				r.Route("/{stackName}", func(r chi.Router) {
					r.Route("/preview", func(r chi.Router) {
						r.Get("/", stackhandler.ExecutePreview())
					})
				})
			})
		})
	})
}

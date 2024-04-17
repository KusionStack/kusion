package route

import (
	"context"
	"expvar"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpswagger "github.com/swaggo/http-swagger"
	"github.com/swaggo/swag/example/basic/docs"
	"kusionstack.io/kusion/pkg/infra/persistence"
	"kusionstack.io/kusion/pkg/server"
	"kusionstack.io/kusion/pkg/server/handler/backend"
	"kusionstack.io/kusion/pkg/server/handler/endpoint"
	"kusionstack.io/kusion/pkg/server/handler/organization"
	"kusionstack.io/kusion/pkg/server/handler/project"
	"kusionstack.io/kusion/pkg/server/handler/source"
	"kusionstack.io/kusion/pkg/server/handler/stack"
	"kusionstack.io/kusion/pkg/server/handler/workspace"
	backendmanager "kusionstack.io/kusion/pkg/server/manager/backend"
	organizationmanager "kusionstack.io/kusion/pkg/server/manager/organization"
	projectmanager "kusionstack.io/kusion/pkg/server/manager/project"
	sourcemanager "kusionstack.io/kusion/pkg/server/manager/source"
	stackmanager "kusionstack.io/kusion/pkg/server/manager/stack"
	workspacemanager "kusionstack.io/kusion/pkg/server/manager/workspace"
	appmiddleware "kusionstack.io/kusion/pkg/server/middleware"

	"kusionstack.io/kusion/pkg/server/util"
)

// NewCoreRoute creates and configures an instance of chi.Mux with the given
// configuration and extra configuration parameters.
func NewCoreRoute(config *server.Config) (*chi.Mux, error) {
	router := chi.NewRouter()

	// Set up middlewares for logging, recovery, and timing, etc.
	router.Use(middleware.RequestID)
	router.Use(appmiddleware.DefaultLogger)
	router.Use(appmiddleware.APILogger)
	router.Use(appmiddleware.Timing)
	router.Use(middleware.Recoverer)

	router.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Set up the API routes for version 1 of the API.
	router.Route("/api/v1", func(r chi.Router) {
		setupRestAPIV1(r, config)
	})

	// Set up the root routes.
	docs.SwaggerInfo.BasePath = "/"
	router.Get("/docs/*", httpswagger.Handler())

	// Endpoint to list all available endpoints in the router.
	router.Get("/endpoints", endpoint.Endpoints(router))

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
	config *server.Config,
) {
	// Set up the logger for the API.
	logger := util.GetLogger(context.TODO())
	logger.Info("Setting up REST API v1...")

	// Set up the persistence layer.
	organizationRepo := persistence.NewOrganizationRepository(config.DB)
	projectRepo := persistence.NewProjectRepository(config.DB)
	stackRepo := persistence.NewStackRepository(config.DB)
	sourceRepo := persistence.NewSourceRepository(config.DB)
	workspaceRepo := persistence.NewWorkspaceRepository(config.DB)
	backendRepo := persistence.NewBackendRepository(config.DB)

	stackManager := stackmanager.NewStackManager(stackRepo, projectRepo, workspaceRepo)
	sourceManager := sourcemanager.NewSourceManager(sourceRepo)
	organizationManager := organizationmanager.NewOrganizationManager(organizationRepo)
	backendManager := backendmanager.NewBackendManager(backendRepo)
	workspaceManager := workspacemanager.NewWorkspaceManager(workspaceRepo, backendRepo)
	projectManager := projectmanager.NewProjectManager(projectRepo, organizationRepo, sourceRepo)

	// Set up the handlers for the resources.
	sourceHandler, err := source.NewHandler(sourceManager)
	if err != nil {
		logger.Error(err, "Error creating source handler...", "error", err)
		return
	}
	orgHandler, err := organization.NewHandler(organizationManager)
	if err != nil {
		logger.Error(err, "Error creating org handler...", "error", err)
		return
	}
	projectHandler, err := project.NewHandler(projectManager)
	if err != nil {
		logger.Error(err, "Error creating project handler...", "error", err)
		return
	}
	stackHandler, err := stack.NewHandler(stackManager)
	if err != nil {
		logger.Error(err, "Error creating stack handler...", "error", err)
		return
	}
	workspaceHandler, err := workspace.NewHandler(workspaceManager)
	if err != nil {
		logger.Error(err, "Error creating workspace handler...", "error", err)
		return
	}
	backendHandler, err := backend.NewHandler(backendManager)
	if err != nil {
		logger.Error(err, "Error creating backend handler...", "error", err)
		return
	}

	// Set up the routes for the resources.
	r.Route("/source", func(r chi.Router) {
		r.Route("/{sourceID}", func(r chi.Router) {
			r.Post("/", sourceHandler.CreateSource())
			r.Get("/", sourceHandler.GetSource())
			r.Put("/", sourceHandler.UpdateSource())
			r.Delete("/", sourceHandler.DeleteSource())
		})
		r.Get("/", sourceHandler.ListSources())
	})
	r.Route("/stack", func(r chi.Router) {
		r.Route("/{stackID}", func(r chi.Router) {
			r.Post("/", stackHandler.CreateStack())
			r.Post("/generate", stackHandler.GenerateStack())
			r.Post("/preview", stackHandler.PreviewStack())
			r.Post("/apply", stackHandler.ApplyStack())
			r.Post("/destroy", stackHandler.DestroyStack())
			r.Get("/", stackHandler.GetStack())
			r.Put("/", stackHandler.UpdateStack())
			r.Delete("/", stackHandler.DeleteStack())
		})
		r.Get("/", stackHandler.ListStacks())
	})
	r.Route("/project", func(r chi.Router) {
		r.Route("/{projectID}", func(r chi.Router) {
			r.Post("/", projectHandler.CreateProject())
			r.Get("/", projectHandler.GetProject())
			r.Put("/", projectHandler.UpdateProject())
			r.Delete("/", projectHandler.DeleteProject())
		})
		r.Get("/", projectHandler.ListProjects())
	})
	r.Route("/org", func(r chi.Router) {
		r.Route("/{organizationID}", func(r chi.Router) {
			r.Post("/", orgHandler.CreateOrganization())
			r.Get("/", orgHandler.GetOrganization())
			r.Put("/", orgHandler.UpdateOrganization())
			r.Delete("/", orgHandler.DeleteOrganization())
		})
		r.Get("/", orgHandler.ListOrganizations())
	})
	r.Route("/workspace", func(r chi.Router) {
		r.Route("/{workspaceID}", func(r chi.Router) {
			r.Post("/", workspaceHandler.CreateWorkspace())
			r.Get("/", workspaceHandler.GetWorkspace())
			r.Put("/", workspaceHandler.UpdateWorkspace())
			r.Delete("/", workspaceHandler.DeleteWorkspace())
		})
		r.Get("/", workspaceHandler.ListWorkspaces())
	})
	r.Route("/backend", func(r chi.Router) {
		r.Route("/{backendID}", func(r chi.Router) {
			r.Post("/", backendHandler.CreateBackend())
			r.Get("/", backendHandler.GetBackend())
			r.Put("/", backendHandler.UpdateBackend())
			r.Delete("/", backendHandler.DeleteBackend())
		})
		r.Get("/", backendHandler.ListBackends())
	})
}

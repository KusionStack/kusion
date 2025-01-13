package route

import (
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpswagger "github.com/swaggo/http-swagger"
	docs "kusionstack.io/kusion/api/openapispec"
	"kusionstack.io/kusion/pkg/infra/persistence"
	"kusionstack.io/kusion/pkg/server"
	"kusionstack.io/kusion/pkg/server/handler/backend"
	"kusionstack.io/kusion/pkg/server/handler/endpoint"
	"kusionstack.io/kusion/pkg/server/handler/module"
	"kusionstack.io/kusion/pkg/server/handler/organization"
	"kusionstack.io/kusion/pkg/server/handler/project"
	"kusionstack.io/kusion/pkg/server/handler/resource"
	"kusionstack.io/kusion/pkg/server/handler/source"
	"kusionstack.io/kusion/pkg/server/handler/stack"
	"kusionstack.io/kusion/pkg/server/handler/workspace"
	backendmanager "kusionstack.io/kusion/pkg/server/manager/backend"
	modulemanager "kusionstack.io/kusion/pkg/server/manager/module"
	organizationmanager "kusionstack.io/kusion/pkg/server/manager/organization"
	projectmanager "kusionstack.io/kusion/pkg/server/manager/project"
	resourcemanager "kusionstack.io/kusion/pkg/server/manager/resource"
	sourcemanager "kusionstack.io/kusion/pkg/server/manager/source"
	stackmanager "kusionstack.io/kusion/pkg/server/manager/stack"
	workspacemanager "kusionstack.io/kusion/pkg/server/manager/workspace"
	appmiddleware "kusionstack.io/kusion/pkg/server/middleware"
	authutil "kusionstack.io/kusion/pkg/server/util/auth"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
	"kusionstack.io/kusion/ui"
)

// NewCoreRoute creates and configures an instance of chi.Mux with the given
// configuration and extra configuration parameters.
func NewCoreRoute(config *server.Config) (*chi.Mux, error) {
	router := chi.NewRouter()
	logger := logutil.GetLogger(context.TODO())

	// Set up middlewares for logging, recovery, and timing, etc.
	router.Use(appmiddleware.TraceID)
	router.Use(appmiddleware.UserID)
	router.Use(appmiddleware.APILoggerMiddleware(config.LogFilePath))
	router.Use(appmiddleware.DefaultLoggerMiddleware(config.LogFilePath))
	router.Use(appmiddleware.Timing)
	router.Use(middleware.Recoverer)

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	// mount pprof routes
	router.Mount("/debug", middleware.Profiler())

	// Set up the API routes for version 1 of the API.
	router.Route("/api/v1", func(r chi.Router) {
		setupRestAPIV1(r, config)
	})

	// Set up the root routes.
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Title = "Kusion Platform Orchestrator API"
	docs.SwaggerInfo.Version = "v0.1.0"
	docs.SwaggerInfo.Description = "Most APIs in Kusion Platform Orchestrator use JWT for authorization.\n" +
		"1. Retrieve a valid token from IAM with the sub claim being either the application name or the user account you are impersonating. \n" +
		"2. Set the token in the Authorization header starting with BEARER \n" +
		"3. The token will be validated against the JWKS endpoint of IAM \n" +
		"4. If the token is valid, the request will be processed. \n" +
		"5. If the token is invalid or missing, the request will be rejected. \n" +
		"6. Also set the x-kusion-user to the caller if impersonation is not enabled in the caller. \n\n" +
		"All responses show below represent the Data field inside the response struct defined across the entire PO. \n"
	router.Get("/docs/*", httpswagger.Handler())

	// Endpoint to list all available endpoints in the router.
	router.Get("/endpoints", endpoint.Endpoints(router))

	// Endpoint to list all available endpoints in the router.
	router.Get("/server-configs", expvar.Handler().ServeHTTP)

	// Endpoint to get server port
	router.Get("/api/server-port", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{
			"port": config.Port,
		})
	})

	// Serve static files from embedded filesystem
	buildFS, _ := fs.Sub(ui.Embedded, "build")
	fileServer := http.FileServer(http.FS(buildFS))

	// Handle static file requests under /public path
	router.Get("/public/*", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/public")
		fileServer.ServeHTTP(w, r)
	})

	// Handle all other paths by serving index.html for SPA routing
	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		indexFile, err := buildFS.Open("index.html")
		if err != nil {
			logger.Warn("Failed to open index.html from embedded filesystem", "error", err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer indexFile.Close()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.Copy(w, indexFile)
	})

	logger.Info(fmt.Sprintf("Listening on :%d", config.Port))
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), router)
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
	logger := logutil.GetLogger(context.TODO())
	logger.Info("Setting up REST API v1...")

	if config.AuthEnabled {
		if len(config.AuthWhitelist) == 0 {
			logger.Info("Auth enabled but whitelist is not set up. Exiting...")
			return
		}
		keyMap, err := authutil.GetJWKSMapFromIAM(context.TODO(), config.AuthKeyType)
		if err != nil {
			logger.Info("Error getting JWKS Map from IAM...")
			return
		}
		r.Use(appmiddleware.TokenAuthMiddleware(keyMap, config.AuthWhitelist, config.LogFilePath))
		logger.Info("Token authorization enabled for REST API v1...")
	}

	// Set up the persistence layer.
	if config.DB != nil && config.AutoMigrate {
		err := persistence.AutoMigrate(config.DB)
		if err != nil {
			logger.Error(err.Error(), "error", "Error auto migrating...")
			return
		}
	}
	organizationRepo := persistence.NewOrganizationRepository(config.DB)
	projectRepo := persistence.NewProjectRepository(config.DB)
	stackRepo := persistence.NewStackRepository(config.DB)
	sourceRepo := persistence.NewSourceRepository(config.DB)
	workspaceRepo := persistence.NewWorkspaceRepository(config.DB)
	backendRepo := persistence.NewBackendRepository(config.DB)
	resourceRepo := persistence.NewResourceRepository(config.DB)
	moduleRepo := persistence.NewModuleRepository(config.DB)
	runRepo := persistence.NewRunRepository(config.DB)

	stackManager := stackmanager.NewStackManager(stackRepo, projectRepo, workspaceRepo, resourceRepo, runRepo, config.DefaultBackend, config.MaxConcurrent)
	sourceManager := sourcemanager.NewSourceManager(sourceRepo)
	organizationManager := organizationmanager.NewOrganizationManager(organizationRepo)
	backendManager := backendmanager.NewBackendManager(backendRepo)
	workspaceManager := workspacemanager.NewWorkspaceManager(workspaceRepo, backendRepo, moduleRepo, config.DefaultBackend)
	projectManager := projectmanager.NewProjectManager(projectRepo, organizationRepo, sourceRepo, config.DefaultSource)
	resourceManager := resourcemanager.NewResourceManager(resourceRepo)
	moduleManager := modulemanager.NewModuleManager(moduleRepo, workspaceRepo, backendRepo)

	// Set up the handlers for the resources.
	sourceHandler, err := source.NewHandler(sourceManager)
	if err != nil {
		logger.Error(err.Error(), "Error creating source handler...", "error", err)
		return
	}
	orgHandler, err := organization.NewHandler(organizationManager)
	if err != nil {
		logger.Error(err.Error(), "Error creating org handler...", "error", err)
		return
	}
	projectHandler, err := project.NewHandler(projectManager)
	if err != nil {
		logger.Error(err.Error(), "Error creating project handler...", "error", err)
		return
	}
	stackHandler, err := stack.NewHandler(stackManager, config.MaxAsyncConcurrent, config.MaxAsyncBuffer)
	if err != nil {
		logger.Error(err.Error(), "Error creating stack handler...", "error", err)
		return
	}
	workspaceHandler, err := workspace.NewHandler(workspaceManager)
	if err != nil {
		logger.Error(err.Error(), "Error creating workspace handler...", "error", err)
		return
	}
	backendHandler, err := backend.NewHandler(backendManager)
	if err != nil {
		logger.Error(err.Error(), "Error creating backend handler...", "error", err)
		return
	}
	resourceHandler, err := resource.NewHandler(resourceManager)
	if err != nil {
		logger.Error(err.Error(), "Error creating resource handler...", "error", err)
		return
	}
	moduleHandler, err := module.NewHandler(moduleManager)
	if err != nil {
		logger.Error(err.Error(), "Error creating module handler", "error", err)
		return
	}

	// Set up the routes for the resources.
	r.Route("/sources", func(r chi.Router) {
		r.Route("/{sourceID}", func(r chi.Router) {
			r.Get("/", sourceHandler.GetSource())
			r.Put("/", sourceHandler.UpdateSource())
			r.Delete("/", sourceHandler.DeleteSource())
		})
		r.Post("/", sourceHandler.CreateSource())
		r.Get("/", sourceHandler.ListSources())
	})
	r.Route("/runs", func(r chi.Router) {
		r.Route("/{runID}", func(r chi.Router) {
			r.Get("/", stackHandler.GetRun())
			r.Get("/result", stackHandler.GetRunResult())
		})
		// r.Post("/", backendHandler.CreateRun())
		r.Get("/", stackHandler.ListRuns())
	})
	r.Route("/stacks", func(r chi.Router) {
		r.Route("/{stackID}", func(r chi.Router) {
			r.Post("/generate", stackHandler.GenerateStack())
			r.Post("/generate/async", stackHandler.GenerateStackAsync())
			r.Post("/preview", stackHandler.PreviewStack())
			r.Post("/preview/async", stackHandler.PreviewStackAsync())
			r.Post("/apply", stackHandler.ApplyStack())
			r.Post("/apply/async", stackHandler.ApplyStackAsync())
			r.Post("/destroy", stackHandler.DestroyStack())
			r.Post("/destroy/async", stackHandler.DestroyStackAsync())
			// r.Route("/variable", func(r chi.Router) {
			// 	r.Post("/", stackHandler.UpdateStackVariable())
			// })
			r.Get("/", stackHandler.GetStack())
			r.Put("/", stackHandler.UpdateStack())
			r.Delete("/", stackHandler.DeleteStack())
		})
		r.Post("/", stackHandler.CreateStack())
		r.Get("/", stackHandler.ListStacks())
	})
	r.Route("/projects", func(r chi.Router) {
		r.Route("/{projectID}", func(r chi.Router) {
			r.Get("/", projectHandler.GetProject())
			r.Put("/", projectHandler.UpdateProject())
			r.Delete("/", projectHandler.DeleteProject())
		})
		r.Post("/", projectHandler.CreateProject())
		r.Get("/", projectHandler.ListProjects())
	})
	r.Route("/orgs", func(r chi.Router) {
		r.Route("/{organizationID}", func(r chi.Router) {
			r.Get("/", orgHandler.GetOrganization())
			r.Put("/", orgHandler.UpdateOrganization())
			r.Delete("/", orgHandler.DeleteOrganization())
		})
		r.Post("/", orgHandler.CreateOrganization())
		r.Get("/", orgHandler.ListOrganizations())
	})
	r.Route("/workspaces", func(r chi.Router) {
		r.Route("/{workspaceID}", func(r chi.Router) {
			// r.Route("/credentials", func(r chi.Router) {
			// 	r.Post("/", workspaceHandler.UpdateWorkspaceCredentials())
			// 	r.Get("/", workspaceHandler.GetWorkspaceCredentials())
			// })
			r.Get("/", workspaceHandler.GetWorkspace())
			r.Put("/", workspaceHandler.UpdateWorkspace())
			r.Delete("/", workspaceHandler.DeleteWorkspace())
			r.Route("/configs", func(r chi.Router) {
				r.Get("/", workspaceHandler.GetWorkspaceConfigs())
				r.Put("/", workspaceHandler.UpdateWorkspaceConfigs())
				r.Route("/mod-deps", func(r chi.Router) {
					r.Post("/", workspaceHandler.CreateWorkspaceModDeps())
				})
			})
		})
		r.Route("/configs", func(r chi.Router) {
			r.Route("/validate", func(r chi.Router) {
				r.Post("/", workspaceHandler.ValidateWorkspaceConfigs())
			})
		})
		r.Post("/", workspaceHandler.CreateWorkspace())
		r.Get("/", workspaceHandler.ListWorkspaces())
	})
	r.Route("/backends", func(r chi.Router) {
		r.Route("/{backendID}", func(r chi.Router) {
			r.Get("/", backendHandler.GetBackend())
			r.Put("/", backendHandler.UpdateBackend())
			r.Delete("/", backendHandler.DeleteBackend())
		})
		r.Post("/", backendHandler.CreateBackend())
		r.Get("/", backendHandler.ListBackends())
	})
	r.Route("/resources", func(r chi.Router) {
		r.Route("/{resourceID}", func(r chi.Router) {
			r.Get("/", resourceHandler.GetResource())
		})
		r.Get("/", resourceHandler.ListResources())
		r.Get("/graph", resourceHandler.GetResourceGraph())
	})
	r.Route("/modules", func(r chi.Router) {
		r.Post("/", moduleHandler.CreateModule())
		r.Get("/", moduleHandler.ListModules())
		r.Route("/{moduleName}", func(r chi.Router) {
			r.Delete("/", moduleHandler.DeleteModule())
			r.Put("/", moduleHandler.UpdateModule())
			r.Get("/", moduleHandler.GetModule())
		})
	})
}

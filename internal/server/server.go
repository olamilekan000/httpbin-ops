package server

import (
	"context"
	"net/http"

	"github.com/TykTechnologies/tyk-devops-assignement/internal/handlers"
	"github.com/TykTechnologies/tyk-devops-assignement/internal/middleware"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	mux        *http.ServeMux
}

// New creates a new Server instance
func New(addr string) *Server {
	mux := http.NewServeMux()
	s := &Server{
		mux: mux,
		httpServer: &http.Server{
			Addr:    addr,
			Handler: middleware.Logging(mux),
		},
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures all the HTTP routes
func (s *Server) setupRoutes() {
	// HTTP method endpoints
	s.mux.HandleFunc("/get", handlers.MethodHandler("GET"))
	s.mux.HandleFunc("/post", handlers.MethodHandler("POST"))
	s.mux.HandleFunc("/put", handlers.MethodHandler("PUT"))
	s.mux.HandleFunc("/patch", handlers.MethodHandler("PATCH"))
	s.mux.HandleFunc("/delete", handlers.MethodHandler("DELETE"))
	s.mux.HandleFunc("/head", handlers.MethodHandler("HEAD"))
	s.mux.HandleFunc("/options", handlers.MethodHandler("OPTIONS"))

	// Utility endpoints
	s.mux.HandleFunc("/headers", handlers.HeadersHandler)
	s.mux.HandleFunc("/ip", handlers.IPHandler)
	s.mux.HandleFunc("/user-agent", handlers.UserAgentHandler)
	s.mux.HandleFunc("/delay/", handlers.DelayHandler)

	// Status code endpoint
	s.mux.HandleFunc("/status/", handlers.StatusHandler)

	// Authentication endpoints
	s.mux.HandleFunc("/basic-auth/", handlers.BasicAuthHandler)
	s.mux.HandleFunc("/bearer", handlers.BearerHandler)
	s.mux.HandleFunc("/digest-auth/", handlers.DigestAuthHandler)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

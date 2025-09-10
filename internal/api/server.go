package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/user/go-web3/internal/config"
)

// Server represents the REST API server
type Server struct {
	router *gin.Engine
	server *http.Server
	config *config.ServerConfig
}

// NewServer creates a new server instance
func NewServer(cfg *config.ServerConfig, handler *Handler) *Server {
	router := gin.Default()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Serve static files
	router.Static("/static", "./static")
	router.StaticFile("/", "./static/index.html")

	// Setup routes
	handler.SetupRoutes(router)

	// Create HTTP server
	addr := cfg.Host + ":" + cfg.Port
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return &Server{
		router: router,
		server: server,
		config: cfg,
	}
}

// Start starts the server
func (s *Server) Start() error {
	log.Printf("Starting server on %s:%s\n", s.config.Host, s.config.Port)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")
	return s.server.Shutdown(ctx)
}

// GracefulShutdown starts a goroutine to handle graceful shutdown
func (s *Server) GracefulShutdown(quit <-chan struct{}) {
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
}

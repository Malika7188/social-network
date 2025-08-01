package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Athooh/social-network/pkg/logger"
)

// Server represents the HTTP server
type Server struct {
	server *http.Server
	logger *logger.Logger
}

// Config holds the server configuration
type Config struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DefaultConfig returns the default server configuration
func DefaultConfig() Config {
	return Config{
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

// New creates a new server instance
func New(config Config, handler http.Handler, logger *logger.Logger) *Server {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return &Server{
		server: server,
		logger: logger,
	}
}

// Start starts the server
func (s *Server) Start() error {
	// Channel for server errors
	errCh := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		s.logger.Info("Starting server on http://%s", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Channel for OS signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either an error or a signal
	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		s.logger.Info("Received signal: %s", sig)
		return s.Shutdown()
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	s.logger.Info("Shutting down server...")

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the server
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("Server shutdown error: %v", err)
		return err
	}

	s.logger.Info("Server shutdown complete")
	return nil
}

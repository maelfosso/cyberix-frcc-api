package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"cyberix.fr/frcc/messaging"
	"cyberix.fr/frcc/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Server struct {
	address  string
	database *storage.Database
	log      *zap.Logger
	mux      chi.Router
	queue    *messaging.Queue
	server   *http.Server
}

type Options struct {
	Database *storage.Database
	Host     string
	Log      *zap.Logger
	Port     int
	Queue    *messaging.Queue
}

func New(opts Options) *Server {
	if opts.Log == nil {
		opts.Log = zap.NewNop()
	}

	address := net.JoinHostPort(opts.Host, strconv.Itoa(opts.Port))
	mux := chi.NewMux()

	return &Server{
		address:  address,
		database: opts.Database,
		log:      opts.Log,
		mux:      mux,
		queue:    opts.Queue,
		server: &http.Server{
			Addr:              address,
			Handler:           mux,
			ReadTimeout:       5 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      5 * time.Second,
			IdleTimeout:       5 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	if err := s.database.Connect(); err != nil {
		return fmt.Errorf("error connection to database: %w", err)
	}

	s.setupRoutes()

	s.log.Info("Starting on", zap.String("address", s.address))
	if err := s.server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error starting server: %w", err)
	}

	return nil
}

func (s *Server) Stop() error {
	s.log.Info("Stopping")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("Error stopping server: %w", err)
	}

	return nil
}

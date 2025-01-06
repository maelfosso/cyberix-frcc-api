package server

import (
	"cyberix.fr/frcc/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (s *Server) setupRoutes() {
	appHandler := handlers.NewAppHandler()

	s.mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	s.mux.Group(func(r chi.Router) {
		appHandler.Health(s.mux)

		r.Route("/auth", func(r chi.Router) {
			appHandler.Register(r, s.database.Storage, s.queue)
			appHandler.Login(r, s.database.Storage, s.queue)
			appHandler.Otp(r, s.database.Storage)
		})

	})
}

package http

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type HealthChecker func(ctx context.Context) error

func NewRouter(handler *UserHandler, jwtSecret string, log zerolog.Logger, dbCheck HealthChecker) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.Recoverer)
	r.Use(RequestID)
	r.Use(Logger(log))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/readyz", func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
		defer cancel()
		if err := dbCheck(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	// Service-to-service: auth-service creates a blank profile on registration.
	r.Post("/internal/profiles", handler.CreateProfileInternal)

	r.Route("/api/users", func(r chi.Router) {
		r.Use(Auth(jwtSecret))
		r.Get("/me", handler.GetMe)
		r.Put("/me", handler.UpdateMe)
		r.Get("/me/addresses", handler.ListAddresses)
		r.Post("/me/addresses", handler.AddAddress)
		r.Delete("/me/addresses/{id}", handler.DeleteAddress)
	})

	return r
}

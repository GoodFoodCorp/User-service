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

func NewRouter(handler *UserHandler, jwtSecret string, log zerolog.Logger, dbCheck HealthChecker, uploadsDir string) http.Handler {
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

	// Public: serves uploaded avatar images (no JWT — plain <img src> tags).
	r.Handle("/api/users/avatars/*", http.StripPrefix("/api/users/avatars/", http.FileServer(http.Dir(uploadsDir))))

	r.Route("/api/users", func(r chi.Router) {
		r.Use(Auth(jwtSecret))
		r.Get("/me", handler.GetMe)
		r.Put("/me", handler.UpdateMe)
		r.Post("/me/avatar", handler.UploadAvatar)
		r.Get("/me/addresses", handler.ListAddresses)
		r.Post("/me/addresses", handler.AddAddress)
		r.Delete("/me/addresses/{id}", handler.DeleteAddress)
		r.Get("/me/favorites", handler.ListFavorites)
		r.Post("/me/favorites", handler.AddFavorite)
		r.Delete("/me/favorites/{kind}/{targetId}", handler.RemoveFavorite)
		r.Get("/me/notifications", handler.GetNotifications)
		r.Put("/me/notifications", handler.UpdateNotifications)
	})

	return r
}

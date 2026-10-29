package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"goodfood/user-service/internal/adapter/avatarstore"
	httpadapter "goodfood/user-service/internal/adapter/http"
	"goodfood/user-service/internal/adapter/postgres"
	"goodfood/user-service/internal/application"
	"goodfood/user-service/internal/config"
	"goodfood/user-service/internal/domain"
)

func main() {
	log := newLogger()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("invalid configuration")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := connectWithRetry(ctx, log, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("database unreachable")
	}
	defer pool.Close()

	if err := postgres.Migrate(cfg.DatabaseURL); err != nil {
		log.Fatal().Err(err).Msg("migrations failed")
	}
	log.Info().Msg("migrations applied")

	var avatars domain.AvatarStore
	localAvatars := ""
	if cfg.S3.Bucket != "" {
		store, err := avatarstore.NewS3Store(avatarstore.S3Config(cfg.S3))
		if err != nil {
			log.Fatal().Err(err).Msg("invalid object storage configuration")
		}
		avatars = store
		log.Info().Str("bucket", cfg.S3.Bucket).Msg("avatars stored in object storage")
	} else {
		avatars = avatarstore.DiskStore{Dir: cfg.UploadsDir, URLPrefix: "/api/users/avatars/"}
		localAvatars = cfg.UploadsDir
		log.Warn().Msg("S3_BUCKET not set: avatars stored on local disk (development only)")
	}

	uc := application.NewUseCases(application.Deps{
		Profiles:      postgres.NewProfileRepository(pool),
		Addresses:     postgres.NewAddressRepository(pool),
		Favorites:     postgres.NewFavoriteRepository(pool),
		Notifications: postgres.NewNotificationPreferencesRepository(pool),
		Avatars:       avatars,
		Eraser:        postgres.NewDataEraser(pool),
	})

	router := httpadapter.NewRouter(httpadapter.NewUserHandler(uc), httpadapter.RouterConfig{
		JWT:             httpadapter.JWTConfig{Secret: cfg.JWTSecret, Issuer: cfg.JWTIssuer, Audience: cfg.JWTAudience},
		ServiceAPIKey:   cfg.ServiceAPIKey,
		Log:             log,
		DBCheck:         func(ctx context.Context) error { return pool.Ping(ctx) },
		LocalAvatarsDir: localAvatars,
	})

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: router, ReadHeaderTimeout: 5 * time.Second}
	log.Info().Str("port", cfg.Port).Msg("user-service started")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("server stopped")
	}
}

func newLogger() zerolog.Logger {
	level, err := zerolog.ParseLevel(strings.ToLower(os.Getenv("LOG_LEVEL")))
	if err != nil || level == zerolog.NoLevel {
		level = zerolog.InfoLevel
	}
	return zerolog.New(os.Stdout).Level(level).With().Timestamp().Str("service", "user-service").Logger()
}

func connectWithRetry(ctx context.Context, log zerolog.Logger, url string) (*pgxpool.Pool, error) {
	for i := 1; ; i++ {
		pool, err := postgres.Connect(ctx, url)
		if err == nil {
			return pool, nil
		}
		if i >= 15 {
			return nil, err
		}
		log.Warn().Err(err).Int("attempt", i).Msg("database not ready, retrying in 2s")
		select {
		case <-time.After(2 * time.Second):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

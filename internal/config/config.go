package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port          string
	Env           string
	DatabaseURL   string
	JWTSecret     string
	JWTIssuer     string
	JWTAudience   string
	ServiceAPIKey string
	LogLevel      string
	UploadsDir    string
	S3            S3
}

// S3 locates the avatars bucket; an empty Bucket keeps avatars on disk.
type S3 struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
	PublicURL string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:          getEnv("PORT", "8087"),
		Env:           getEnv("ENV", "development"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTIssuer:     getEnv("JWT_ISSUER", "goodfood-auth"),
		JWTAudience:   getEnv("JWT_AUDIENCE", "goodfood-api"),
		ServiceAPIKey: os.Getenv("SERVICE_API_KEY"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		UploadsDir:    getEnv("UPLOADS_DIR", "/data/uploads"),
		S3: S3{
			Endpoint:  os.Getenv("S3_ENDPOINT"),
			Bucket:    os.Getenv("S3_BUCKET"),
			AccessKey: os.Getenv("S3_ACCESS_KEY"),
			SecretKey: os.Getenv("S3_SECRET_KEY"),
			UseSSL:    getEnv("S3_USE_SSL", "true") == "true",
			PublicURL: os.Getenv("S3_PUBLIC_URL"),
		},
	}
	switch {
	case cfg.DatabaseURL == "":
		return nil, fmt.Errorf("DATABASE_URL is required")
	case cfg.JWTSecret == "":
		return nil, fmt.Errorf("JWT_SECRET is required")
	case cfg.ServiceAPIKey == "":
		return nil, fmt.Errorf("SERVICE_API_KEY is required")
	case cfg.S3.Bucket != "" && (cfg.S3.Endpoint == "" || cfg.S3.PublicURL == ""):
		return nil, fmt.Errorf("S3_ENDPOINT and S3_PUBLIC_URL are required with S3_BUCKET")
	case cfg.Env == "production" && cfg.S3.Bucket == "":
		return nil, fmt.Errorf("S3_BUCKET is required in production: avatars on a pod disk are lost and not shared")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

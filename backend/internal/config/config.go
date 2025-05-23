package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string
	DatabaseURL string
}

func NewConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, using environment variables")
	}

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	env := os.Getenv("ENV")

	if env == "" {
		env = "DEVELOPMENT"
	}

	return &Config{DatabaseURL: databaseURL, Env: env}, nil
}

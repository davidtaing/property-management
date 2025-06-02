package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/davidtaing/property-management/api"
	"github.com/davidtaing/property-management/internal/config"
	"github.com/davidtaing/property-management/internal/middleware"
	oapifilter "github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	oapiMiddleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/rs/cors"
)

func main() {
	logger := setupLogger()
	config, err := config.NewConfig()

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	clerk.SetKey(config.ClerkKey)

	dbpool, err := pgxpool.New(context.Background(), config.DatabaseURL)

	if err != nil {
		logger.Error("Unable to connect to the database")
		os.Exit(1)
	}

	defer dbpool.Close()

	err = dbpool.Ping(context.Background())

	if err != nil {
		logger.Error("Unable to ping the database")
		os.Exit(1)
	}

	logger.Info("Connected to database")

	swagger, err := api.GetSwagger()
	if err != nil {
		msg := fmt.Sprintf("Error loading swagger spec:\n %s", err)
		logger.Error(msg)
		os.Exit(1)
	}

	swagger.Servers = nil

	server := api.NewServer(dbpool, logger)

	r := mux.NewRouter()

	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(middleware.AuthMiddleware())

	validatorOptions := &oapiMiddleware.Options{}

	// stub this out and use the clerkhttp middleware instead
	validatorOptions.Options.AuthenticationFunc = func(c context.Context, input *oapifilter.AuthenticationInput) error {
		return nil
	}

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	r.Use(oapiMiddleware.OapiRequestValidatorWithOptions(swagger, validatorOptions))

	h := api.HandlerFromMux(server, r)

	c := setupCors(config, logger)
	h = clerkhttp.WithHeaderAuthorization()(h)
	h = c.Handler(h)

	s := &http.Server{
		Handler: h,
		Addr:    "0.0.0.0:8080",
	}

	logger.Info("Starting server on port 8080")

	log.Fatal(s.ListenAndServe())
}

func setupLogger() *slog.Logger {
	level := slog.LevelDebug

	if os.Getenv("ENV") == "PRODUCTION" {
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	return logger
}

func setupCors(config *config.Config, logger *slog.Logger) *cors.Cors {
	allowedOrigins := []string{"*"}

	if config.Env == "PRODUCTION" {
		allowedOrigins = []string{"https://davidtaing.github.io"}
	}

	logger.Debug("Allowed origins", "origins", allowedOrigins)

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Accept"},
		ExposedHeaders:   []string{"Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	return c
}

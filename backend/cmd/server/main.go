package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/davidtaing/property-management/api"
	"github.com/davidtaing/property-management/internal/config"
	"github.com/davidtaing/property-management/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	oapiMiddleware "github.com/oapi-codegen/nethttp-middleware"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	config, err := config.NewConfig()

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

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

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	r.Use(oapiMiddleware.OapiRequestValidator(swagger))

	r.Use(middleware.LoggingMiddleware(logger))

	h := api.HandlerFromMux(server, r)

	s := &http.Server{
		Handler: h,
		Addr:    "0.0.0.0:8080",
	}

	logger.Info("Starting server on port 8080")

	log.Fatal(s.ListenAndServe())
}

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/davidtaing/property-management/api"
	"github.com/davidtaing/property-management/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	oapiMiddleware "github.com/oapi-codegen/nethttp-middleware"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dbpool, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5432/property_management")

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

	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	server := api.NewServer(dbpool, logger)

	r := mux.NewRouter()

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	r.Use(oapiMiddleware.OapiRequestValidator(swagger))

	// Add the logging middleware to the router
	r.Use(middleware.LoggingMiddleware(logger))

	// get an `http.Handler` that we can use
	h := api.HandlerFromMux(server, r)

	s := &http.Server{
		Handler: h,
		Addr:    "0.0.0.0:8080",
	}

	// fmt.Println("Starting server on port 8080")
	logger.Info("Starting server on port 8080")

	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/davidtaing/property-management/api"
)

func main() {
	dbpool, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5432/property_management")
	if err != nil {
		fmt.Println("Unable to connect to database:", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	server := api.NewServer(dbpool)

	r := mux.NewRouter()

	// get an `http.Handler` that we can use
	h := api.HandlerFromMux(server, r)

	s := &http.Server{
		Handler: h,
		Addr:    "0.0.0.0:8080",
	}

	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}
# Property Management

This project is a full-stack property management application designed to streamline the processes of managing rental properties, tenants, leases. It aims to provide property managers with a centralized, user-friendly platform for handling day-to-day operations efficiently.

## Features (Planned)
- RESTful API for property, tenant, and lease management
- Authentication
- Payment tracking and invoicing

## Tech Stack

### Schema
- TypeSpec - defines the REST API schema using a TypeScript like syntax. This serves as the single source of truth for the API and is used to generate an OpenAPI specification.

### Frontend
- TBD

### Backend
- Golang
- PostgreSQL
- oapi-codegen - Automatically generates Go server stubs and type-safe HTTP handlers from the OpenAPI schema.
- goose -  Handles database migrations for version-controlled schema updates.
- gorilla/mux -  A powerful HTTP router and URL matcher for building REST APIs in Go.


start-server:
	cd backend && go run cmd/server/main.go

gen-openapi:
	tsp compile --output-dir . typespec/main.tsp

gen-backend:
	cd backend/api && go generate

setup-local-db:
	psql postgres://postgres:postgres@localhost:5432 -c "CREATE DATABASE property_management"

migrate:
	@goose -dir migrations postgres $(DATABASE_URL) up

migrate-status:
	@goose -dir migrations postgres $(DATABASE_URL) status

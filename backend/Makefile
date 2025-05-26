start-server:
	go run cmd/server/server.go

setup-local-db:
	psql postgres://postgres:postgres@localhost:5432 -c "CREATE DATABASE property_management"

migrate:
	@goose -dir migrations postgres $(DATABASE_URL) up

migrate-status:
	@goose -dir migrations postgres $(DATABASE_URL) status

gen-openapi:
	tsp compile typespec/main.tsp

gen-backend:
	cd backend/api && go generate

setup-db:
	psql postgres://postgres:postgres@localhost:5432 -c "CREATE DATABASE property_management"

migrations-up:
	goose -dir migrations postgres "postgres://postgres:postgres@localhost:5432/property_management" up

migrations-status:
	goose -dir migrations postgres "postgres://postgres:postgres@localhost:5432/property_management" status


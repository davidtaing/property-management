gen-openapi:
	tsp compile --output-dir . typespec/main.tsp

gen-backend:
	cd backend/api && go generate
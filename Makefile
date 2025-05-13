gen-openapi:
	tsp compile typespec/main.tsp

gen-backend:
	cd backend/api && go generate
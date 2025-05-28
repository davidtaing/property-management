generate:
	make gen-openapi && make gen-backend

gen-openapi:
	tsp compile --output-dir . typespec/main.tsp

gen-backend:
	cd backend/api && go generate
gen-openapi:
	tsp compile typespec/main.tsp

gen-backend:
	cd backend && oapi-codegen -config ../oapi-codegen-config.yaml ../tsp-output/schema/openapi.yaml
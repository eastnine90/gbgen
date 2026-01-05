package growthbookapi

//go:generate go tool oapi-codegen -include-operation-ids listFeatures -package growthbookapi -generate types,client -o client.go $GEN_OPENAPI_FILE

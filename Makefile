
.PHONY:
build-spec-tests:
	go run sszgen/*.go --path ./spectests/structs.go --include ./spectests/external

.PHONY:
ef-tests:
	./scripts/download-ef-tests.sh v0.10.0

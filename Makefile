
.PHONY:
build-spec-tests:
	go run sszgen/*.go --path ./spectests/structs.go --include ./spectests/external

.PHONY:
get-spec-tests:
	./scripts/download-spec-tests.sh v0.10.0

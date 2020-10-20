
.PHONY:
build-spec-tests:
	go run sszgen/*.go --path ./spectests/structs.go --include ./spectests/external,./spectests/external2

.PHONY:
get-spec-tests:
	./scripts/download-spec-tests.sh v0.10.0

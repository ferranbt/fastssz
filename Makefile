
.PHONY:
build-spec-tests:
	go run sszgen/*.go --path ./spectests/structs.go --include ./spectests/external,./spectests/external2

build-spec-tests-tree:
	go run sszgen/*.go --path ./spectests/structs.go --objs AttestationData --experimental

.PHONY:
get-spec-tests:
	./scripts/download-spec-tests.sh v0.10.0

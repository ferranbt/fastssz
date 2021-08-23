
.PHONY:
build-spec-tests:
	go run github.com/ferranbt/fastssz/sszgen --path ./spectests/phase0/structs.go --include ./spectests/external,./spectests/external2
	go run github.com/ferranbt/fastssz/sszgen --path ./spectests/phase0/minimal/structs.go --include ./spectests/external,./spectests/external2
	go run github.com/ferranbt/fastssz/sszgen --path ./spectests/altair/structs.go --include ./spectests/external,./spectests/external2
	go run github.com/ferranbt/fastssz/sszgen --path ./spectests/altair/minimal/structs.go --include ./spectests/external,./spectests/external2

build-spec-tests-tree:
	go run github.com/ferranbt/fastssz/sszgen --path ./spectests/structs.go --objs AttestationData --experimental

.PHONY:
get-spec-tests:
	./scripts/download-spec-tests.sh v0.10.0

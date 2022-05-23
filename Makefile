
.PHONY:
build-spec-tests:
	go run github.com/ferranbt/fastssz/sszgen --path ./spectests/structs.go --exclude-objs Hash --include ./spectests/external,./spectests/external2 --experimental

build-spec-tests-tree:
	go run github.com/ferranbt/fastssz/sszgen --path ./spectests/structs.go --objs AttestationData --experimental

.PHONY:
get-spec-tests:
	./scripts/download-spec-tests.sh v1.1.10

.PHONY:
generate-testcases:
	cd sszgen/testcases && go generate

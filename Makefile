
.PHONY:
build-spec-tests:
	go run github.com/ferranbt/fastssz/sszgen --path ./spectests/structs.go --include ./spectests/external,./spectests/external2 --exclude-objs Hash
	go run github.com/ferranbt/fastssz/sszgen --path ./tests/codetrie.go

.PHONY:
get-spec-tests:
	./scripts/download-spec-tests.sh v1.1.10

.PHONY:
generate-testcases:
	cd sszgen/testcases && go generate

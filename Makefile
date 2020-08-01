
.PHONY:
build-spec-tests:
	go run sszgen/*.go --path ./spectests/structs.go

.PHONY:
ef-tests:
	./scripts/download-ef-tests.sh v0.10.0

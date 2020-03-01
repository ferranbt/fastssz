
.PHONY:
build-spec-tests:
	go run sszgen/main.go --path ./spectests/structs.go


.PHONY:
build-spec-tests:
	go run sszgen/*.go --path ./spectests/structs.go

gen:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.1
	cd shared/proto && protoc --go_out=../ --go_opt=paths=source_relative game.proto

.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run

.PHONY: test
test:
	go test -v -race ./...

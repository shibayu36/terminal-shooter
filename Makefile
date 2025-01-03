gen:
	cd shared/proto && protoc --go_out=../ --go_opt=paths=source_relative game.proto

.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run

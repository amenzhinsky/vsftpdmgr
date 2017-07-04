all:
	@go build -i -ldflags="-s -w"

test:
	@go test -i -race ./...
	@go test -v -race -cover $(shell go list ./... | grep -v vendor)

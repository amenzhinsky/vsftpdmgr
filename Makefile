build:
	@go build -i -ldflags="-s -w"

release: build
	@tar cvzf vsftpdmgr_linux_amd64.tar.gz vsftpdmgr

test:
	@go test -i -race ./...
	@go test -v -race -cover $(shell go list ./... | grep -v vendor)

todo: # SKIPTODO
	grep -rni . | grep -i todo: | grep -vi SKIPTODO

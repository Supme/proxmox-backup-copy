.PHONY: build
build:
	go build -ldflags "-s -w" ./

.PHONY: test
test:
	go test -v -race -timeout 30s ./...

.PHONY: cover
cover:
	go test -cover ./...

.DEFAULT_GOAL := build

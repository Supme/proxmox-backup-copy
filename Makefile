.PHONY: build
build:
	go build -v ./

.PHONY: test
test:
	go test -v -race -timeout 30s ./...

.PHONY: cover
cover:
	go test -cover ./...

.DEFAULT_GOAL := build

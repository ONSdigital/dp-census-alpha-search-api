SHELL=bash

BUILD=build
BIN_DIR?=.

SEARCH_API=search

build:
	go generate ./...
	@mkdir -p $(BUILD)/$(BIN_DIR)
	go build -o $(BUILD)/$(BIN_DIR)/$(SEARCH_API) cmd/$(SEARCH_API)/main.go

debug: build
	HUMAN_LOG=1 go run -race cmd/$(SEARCH_API)/main.go

test:
	go test -cover -race ./...

.PHONY: build api test

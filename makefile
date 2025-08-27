# Makefile at repo root

BINARY_NAME ?= git-semver
BUILD_DIR := cli
BIN_DIR := bin

.PHONY: all build clean test

all: build

## Build the CLI into ./bin/
build:
	@echo "Building $(BINARY_NAME) from $(BUILD_DIR)..."
	@mkdir -p $(BIN_DIR)
	@cd $(BUILD_DIR) && go build -o ../$(BIN_DIR)/$(BINARY_NAME)

## Run all Go tests
test:
	@echo "Running tests..."
	@go test ./...

## Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)/

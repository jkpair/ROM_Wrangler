APP_NAME := romwrangler
BUILD_DIR := bin
GO := go

.PHONY: all build run test lint clean

all: build

build:
	$(GO) build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/romwrangler

run:
	$(GO) run ./cmd/romwrangler

test:
	$(GO) test ./... -v

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)

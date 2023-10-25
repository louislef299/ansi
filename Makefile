.DEFAULT_GOAL := default
.PHONY: test lint

default: lint test test-colors

test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

test-colors:
	@echo "Running tests for output colors"
	@go test -v -run TestBufferStagesColor
	@go test -v -run TestStandardStagesColor

lint:
	@echo "Linting Go program files"
	@golangci-lint run
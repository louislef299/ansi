.DEFAULT_GOAL := default
.PHONY: test lint

default: lint test

test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

lint:
	@echo "Linting Go program files"
	@golangci-lint run
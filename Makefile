test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

lint:
	@echo "Linting Go program files"
	@golangci-lint run

# Command for running the linter.
lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

# Command for running tests with coverage.
test:
	@echo "Running tests..."
	@go test ./... -coverprofile=coverage.out

# Command for showing test coverage.
coverage:
	@echo "Displaying test coverage..."
	@go tool cover -func=coverage.out

# Command for running the backend.
run-backend:
	@echo "Starting the backend..."
	@go run ./cmd/main.go


.PHONY: lint test coverage run-backend clean

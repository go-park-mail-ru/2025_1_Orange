# Команда для запуска линтера.
lint:
	@echo "Запуск golangci-lint..."
	@golangci-lint run ./...

# Команда для запуска тестов с покрытием.
test:
	@echo "Запуск тестов..."
	@go test ./... -coverprofile=coverage.out

# Команда для отображения покрытия тестов.
coverage:
	@echo "Отображение покрытия тестов..."
	@go tool cover -func=coverage.out

# Команда для запуска бэкенда.
run-backend:
	@echo "Запуск бэкенда..."
	@go run ./cmd/main.go


.PHONY: lint test coverage run-backend clean

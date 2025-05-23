up-build:
	docker-compose up --build

up:
	docker-compose up

lint:
	golangci-lint run

.PHONY: easyjson
easyjson:
	@echo "Генерация easyjson..."
	@echo "Обработка internal/entity/dto..."
	@for file in $$(find ./internal/entity/dto -name '*.go' | grep -v "_easyjson.go"); do \
		easyjson -all $$file; \
	done
	@echo "Генерация завершена"
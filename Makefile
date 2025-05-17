up-build:
	docker-compose up --build

up:
	docker-compose up

lint:
	golangci-lint run
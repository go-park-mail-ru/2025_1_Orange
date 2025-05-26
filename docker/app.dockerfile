FROM golang:1.23-alpine AS builder

WORKDIR /app

# Установка зависимостей проекта
COPY go.mod ./
COPY go.sum ./

# Установка мигратора БД
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

#COPY . .
COPY .env ./
COPY configs/main.yml ./configs/main.yml
COPY db/migrations ./db/migrations
COPY docs ./docs
COPY cmd/app/main.go ./cmd/app/main.go
COPY internal ./internal
COPY pkg ./pkg
COPY static/templates ./static/templates

RUN go mod download

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o ./.bin ./cmd/app/main.go

# Запуск (используем минимальный по весу образ Linux)
FROM alpine:3.18
WORKDIR /app

# Копируем бинарник, миграции и migrate
COPY --from=builder /app/.env .
COPY --from=builder /app/configs/main.yml ./configs/main.yml
COPY --from=builder /app/.bin .
COPY --from=builder /app/db/migrations ./migrations
COPY --from=builder /go/bin/migrate .
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/static/templates ./static/templates

RUN apk add --no-cache tzdata libc6-compat
ENV TZ=Europe/Moscow

# Директория для аватарок
RUN mkdir -p /app/assets


# Сборка (используем минимальный по весу Linux с Go)
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Установка зависимостей проекта
COPY go.mod .
COPY go.sum .
RUN go mod download

# Установка мигратора БД
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY . .

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

# Директория для аватарок
RUN mkdir -p /app/assets

# tzdata - указываем, какой часовой пояс использовать
# libc6-compat - для migrate
RUN apk add --no-cache tzdata libc6-compat
ENV TZ=Europe/Moscow

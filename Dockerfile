# Сборка (используем минимальный по весу Linux с Go)
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .

# Надо установить зависимости из go.mod, загрузить либу для миграций,
# затем собрать бинарник (т.к. это экономит память)
RUN go mod download && \
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest && \
    CGO_ENABLED=0 GOOS=linux go build -o ./.bin ./cmd/app/main.go  # Бинарник в ./.bin

# Запуск (используем минимальный по весу образ Linux)
FROM alpine:3.18
WORKDIR /app

# Копируем бинарник, миграции и migrate
COPY --from=builder /app/.env .
COPY --from=builder /app/configs/main.yml ./configs/main.yml
COPY --from=builder /app/.bin .
COPY --from=builder /app/db/migrations ./migrations
COPY --from=builder /go/bin/migrate .


# tzdata - указываем, какой часовой пояс использовать
# libc6-compat - для migrate
RUN apk add --no-cache tzdata libc6-compat
ENV TZ=Europe/Moscow

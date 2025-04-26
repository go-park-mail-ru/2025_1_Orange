FROM golang:1.23-alpine AS builder

WORKDIR /app

# Установка зависимостей
COPY go.mod .
COPY go.sum .
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o poll ./cmd/poll/main.go

# Финальный образ
FROM alpine:3.18
WORKDIR /app

# Копируем необходимые файлы
COPY --from=builder /app/.env .
COPY --from=builder /app/configs/poll.yml ./configs/poll.yml
COPY --from=builder /app/poll .

# Установка зависимостей
RUN apk add --no-cache tzdata libc6-compat
ENV TZ=Europe/Moscow

RUN chmod +x /app/poll

ENTRYPOINT ["./poll"]

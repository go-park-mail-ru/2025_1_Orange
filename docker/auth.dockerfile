FROM golang:1.23-alpine AS builder

WORKDIR /auth

COPY go.mod ./
COPY go.sum ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./.bin ./cmd/auth/main.go

FROM alpine:3.18
WORKDIR /auth

COPY --from=builder /auth/.env .env
COPY --from=builder /auth/configs/auth.yml ./configs/auth.yml
COPY --from=builder /auth/.bin .

RUN apk add --no-cache tzdata libc6-compat
ENV TZ=Europe/Moscow

CMD ["./.bin"]

FROM golang:1.23-alpine AS builder

WORKDIR /static

COPY go.mod ./
COPY go.sum ./

COPY .env ./
COPY cmd/static/main.go ./cmd/static/main.go
COPY internal ./internal
COPY pkg ./pkg
COPY configs/static.yml ./configs/static.yml

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o ./.bin ./cmd/static/main.go

FROM alpine:3.18
WORKDIR /static

COPY --from=builder /static/.env .env
COPY --from=builder /static/configs/static.yml ./configs/static.yml
COPY --from=builder /static/.bin .

CMD ["./.bin"]

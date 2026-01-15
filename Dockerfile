FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
RUN go build -o main ./cmd/main

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations
COPY .env /app/
ENV DB_HOST=db

ENV GOOSE_DRIVER=postgres
ENV GOOSE_DBSTRING=postgres://postgres:postgres@db:5432/incidents_service?sslmode=disable
ENV GOOSE_MIGRATION_DIR=./migrations
COPY --from=builder /go/bin/goose /usr/local/bin/goose
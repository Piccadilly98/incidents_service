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
ENV DB_PORT=5432
ENV SERVER_ADDR=0.0.0.0      
ENV SERVER_PORT=8080 
ENV GOOSE_DBSTRING=postgres://postgres:postgres@db:5432/incidents_service 
ENV REDIS_ADDR=cache:6379
COPY --from=builder /go/bin/goose /usr/local/bin/goose
EXPOSE 8080
COPY entrypoint.sh /app/
RUN chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]

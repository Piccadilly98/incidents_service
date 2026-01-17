all: goose-up start

goose-up:
	goose up
goose-down:
	goose down

start:
	go run cmd/main/main.go

tests-cover:
	go test -cover ./...

tests-v:
	go test -v ./...

tests-all:
	go test -v -cover ./...
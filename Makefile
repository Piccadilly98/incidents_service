all: goose-up start

goose-up:
	goose up
goose-down:
	goose down

start:
	go run cmd/main/main.go
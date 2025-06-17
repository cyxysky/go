.PHONY: help build run clean test docker-up docker-down

help:
	@echo "Available commands:"
	@echo "  make build       - Build the application"
	@echo "  make run         - Run the application"
	@echo "  make test        - Run tests"
	@echo "  make clean       - Clean build files"
	@echo "  make docker-up   - Start database containers"
	@echo "  make docker-down - Stop database containers"
	@echo "  make deps        - Download dependencies"

build:
	go build -o bin/gin-web-api main.go

run:
	go run main.go

deps:
	go mod download
	go mod tidy

test:
	go test -v ./...

clean:
	rm -rf bin/
	go clean

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down 
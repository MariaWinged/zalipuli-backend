BINARY   := server
MODULE   := zalipuli
IMAGE    := $(MODULE)

.PHONY: all build run generate test lint tidy clean docker-build docker-run compose-up compose-down

all: generate build

## Build the binary
build:
	go build -o $(BINARY) .

## Run the server locally
run:
	go run .

## Run code generation (oapi-codegen)
generate:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=./api/gen-server.yaml ./api/openapi.yaml && \
    go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=./api/gen-client.yaml ./api/openapi.yaml && \
    go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=./api/gen-types.yaml ./api/openapi.yaml

## Run all tests
test:
	go test ./...

## Run tests with verbose output
test-v:
	go test -v ./...

## Run tests with race detector
test-race:
	go test -race ./...

## Tidy and verify dependencies
tidy:
	go mod tidy
	go mod verify

## Remove built binary
clean:
	rm -f $(BINARY)

## Build Docker image
docker-build:
	docker build -t $(IMAGE) .

## Run Docker container (reads PORT from env, defaults to 8080)
docker-run:
	docker run --rm -p $${PORT:-8080}:$${PORT:-8080} -e PORT=$${PORT:-8080} $(IMAGE)

## Start docker-compose stack
compose-up:
	docker-compose up -d

## Stop docker-compose stack
compose-down:
	docker-compose down

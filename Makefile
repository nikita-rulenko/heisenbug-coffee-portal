.PHONY: run build test test-unit test-integration test-api test-e2e lint clean seed

run:
	go run ./cmd/server

build:
	go build -o bin/portal ./cmd/server

test:
	go test -v -race -count=1 ./...

test-unit:
	go test -v -race -count=1 -run TestUnit ./...

test-integration:
	go test -v -race -count=1 -run TestIntegration ./...

test-api:
	go test -v -race -count=1 -run TestAPI ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ data/

seed:
	go run ./cmd/server -seed-only

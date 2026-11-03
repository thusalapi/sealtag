build:
	go build -o bin/sealtag ./cmd/sealtag

test:
	go test ./...

lint:
	go vet ./...

.PHONY: build test lint

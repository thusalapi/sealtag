build:
	go build -o bin/shipwright ./cmd/shipwright

test:
	go test ./...

lint:
	go vet ./...

.PHONY: build test lint

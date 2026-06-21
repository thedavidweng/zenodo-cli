.PHONY: fmt vet test build lint

fmt:
	gofmt -w ./cmd ./internal

vet:
	go vet ./...

test:
	go test ./...

build:
	go build -o bin/zenodo ./cmd/zenodo

lint: fmt vet test
	golangci-lint run

.PHONY: test lint tidy build

test:
	go test -v -count=1 ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

build:
	go build ./...

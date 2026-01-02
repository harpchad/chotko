.PHONY: all build test lint fmt clean update

BINARY := chotko
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

all: fmt lint test build

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/$(BINARY)

test:
	go test -race -cover ./...

lint:
	golangci-lint run ./...

fmt:
	gofumpt -w .

clean:
	rm -f $(BINARY)

update:
	go get -u ./...
	go mod tidy
	pre-commit autoupdate

GO := go
BINARY := jotform
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X github.com/jotform/jotform-cli/cmd.Version=$(VERSION) \
	-X github.com/jotform/jotform-cli/cmd.Commit=$(COMMIT) \
	-X github.com/jotform/jotform-cli/cmd.BuildDate=$(DATE)

.PHONY: build test lint clean install

build:
	$(GO) build -ldflags="$(LDFLAGS)" -o $(BINARY) .

test:
	$(GO) test ./... -race -coverprofile=coverage.out
	@echo "Coverage report: go tool cover -html=coverage.out"

lint:
	$(GO) vet ./...
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"

clean:
	rm -f $(BINARY) coverage.out

install: build
	cp $(BINARY) /usr/local/bin/$(BINARY)

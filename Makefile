GO := go
BINARY := jotform
ALIAS := jf
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X github.com/jotform/jotform-cli/cmd.Version=$(VERSION) \
	-X github.com/jotform/jotform-cli/cmd.Commit=$(COMMIT) \
	-X github.com/jotform/jotform-cli/cmd.BuildDate=$(DATE)

.PHONY: build test lint clean install install-jf dev

build:
	$(GO) build -ldflags="$(LDFLAGS)" -o $(BINARY) .

# install: builds and installs both 'jotform' and 'jf' symlink to /usr/local/bin
install: build
	cp $(BINARY) /usr/local/bin/$(BINARY)
	ln -sf /usr/local/bin/$(BINARY) /usr/local/bin/$(ALIAS)
	@echo "✓ Installed: jotform + jf → /usr/local/bin/"

# install-jf: adds jf symlink only (if jotform is already installed)
install-jf:
	@which jotform > /dev/null 2>&1 || (echo "❌ jotform not found, run 'make install' first" && exit 1)
	ln -sf $$(which jotform) /usr/local/bin/$(ALIAS)
	@echo "✓ jf → $$(which jotform)"

test:
	$(GO) test ./... -race -coverprofile=coverage.out
	@echo "Coverage report: go tool cover -html=coverage.out"

lint:
	$(GO) vet ./...
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"

clean:
	rm -f $(BINARY) coverage.out

# dev: run with hot-reload awareness (no symlink, just local binary)
dev: build
	@echo "Built ./$(BINARY) — run as: ./$(BINARY) or alias jf=./$(BINARY)"

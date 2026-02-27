APP       := rto
MODULE    := rto
GOFLAGS   ?=
GO_LDFLAGS ?=

.PHONY: all build release install run clean fmt vet lint test test-v test-race coverage tidy help

all: fmt vet test build  ## Default: format, vet, test, build

build:  ## Compile the binary
	go build $(GOFLAGS) -ldflags "$(GO_LDFLAGS)" -o $(APP) .

release:  ## Optimized release build (stripped, no debug info)
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w $(GO_LDFLAGS)" -o $(APP) .

install:  ## Install optimized release build to $GOPATH/bin
	CGO_ENABLED=0 go install -trimpath -ldflags "-s -w $(GO_LDFLAGS)" .

run: build  ## Build and launch the TUI
	./$(APP)

clean:  ## Remove build artifacts and coverage files
	rm -f $(APP) coverage.out coverage.html

fmt:  ## Format all Go source files
	go fmt ./...

vet:  ## Run go vet
	go vet ./...

lint:  ## Run staticcheck (install: go install honnef.co/go/tools/cmd/staticcheck@latest)
	staticcheck ./...

test:  ## Run all tests
	go test ./...

test-v:  ## Run all tests with verbose output
	go test -v ./...

test-race:  ## Run tests with race detector
	go test -race ./...

coverage:  ## Generate HTML coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Open coverage.html in your browser"

tidy:  ## Tidy and verify module dependencies
	go mod tidy
	go mod verify

help:  ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

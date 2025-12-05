.PHONY: build run clean help

BINARY_NAME=git-download-stats
GO=go

help:
	@echo "Available targets:"
	@echo "  make build       - Build the Go binary"
	@echo "  make run         - Run the program (requires -owner and -repo)"
	@echo "  make test        - Run tests"
	@echo "  make clean       - Remove built binaries"
	@echo "  make deps        - Download dependencies"

deps:
	$(GO) mod download
	$(GO) mod verify

build: deps
	$(GO) build -o $(BINARY_NAME) -v

run: build
	./$(BINARY_NAME) $(ARGS)

clean:
	$(GO) clean
	rm -f $(BINARY_NAME)

test:
	$(GO) test -v ./...

BINARY    := stacklit
CMD       := ./cmd/stacklit
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -ldflags "-X github.com/glincker/stacklit/internal/cli.Version=$(VERSION)"

.PHONY: build run test clean

build:
	go build $(LDFLAGS) -o $(BINARY) $(CMD)

run: build
	./$(BINARY)

test:
	go test ./...

clean:
	rm -f $(BINARY)
	rm -rf dist/

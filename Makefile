# Makefile for ezysearch

# Variables
BINARY=ezysearch
VERSION=1.0.0
LDFLAGS=-s -w -X github.com/tumillanino/ezysearch/internal/cli.Version=${VERSION}

# Build for current platform
build:
	go build -ldflags "${LDFLAGS}" -o ${BINARY} .

# Install to system
install: build
	install -Dm755 ${BINARY} ${DESTDIR}/usr/local/bin/${BINARY}

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o bin/${BINARY}-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o bin/${BINARY}-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o bin/${BINARY}-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o bin/${BINARY}-darwin-arm64 .

# Clean build artifacts
clean:
	rm -f ${BINARY}
	rm -rf bin/

# Run tests
test:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run all checks
check: fmt vet test

.PHONY: build install build-all clean test fmt vet check

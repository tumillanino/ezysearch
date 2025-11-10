# Makefile for ezysearch

# Variables
BINARY=ezysearch
VERSION=1.0.0
MAIN_DIR=cmd/ezysearch

# Build for current platform
build:
	go build -o ${BINARY} ${MAIN_DIR}/main.go

# Install to system
install: build
	install -Dm755 ${BINARY} ${DESTDIR}/usr/local/bin/${BINARY}

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o bin/${BINARY}-linux-amd64 ${MAIN_DIR}/main.go
	GOOS=linux GOARCH=arm64 go build -o bin/${BINARY}-linux-arm64 ${MAIN_DIR}/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/${BINARY}-darwin-amd64 ${MAIN_DIR}/main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/${BINARY}-darwin-arm64 ${MAIN_DIR}/main.go

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
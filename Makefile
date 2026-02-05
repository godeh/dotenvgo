.PHONY: all test coverage audit fmt tidy

all: fmt audit test

# Run tests
test:
	go test -v -race ./...

# Run tests with coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Audit code (vet and staticcheck)
audit:
	go vet ./...
	# Check if staticcheck is installed, if not, print warning
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not found, skipping (install with 'go install honnef.co/go/tools/cmd/staticcheck@latest')"; \
	fi

# Format code
fmt:
	go fmt ./...

# Tidy dependencies
tidy:
	go mod tidy
	go mod verify

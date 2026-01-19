.PHONY: dev build run test clean install-air

# Development
dev: install-air
	air

# Build
build:
	go build -o bin/app main.go

# Run production build
run: build
	./bin/app

# Test
test:
	go test ./...

test-cover:
	go test -cover ./...

# Install Air for hot reloading
install-air:
	@which air > /dev/null || go install github.com/cosmtrek/air@latest

# Install dependencies
deps:
	go mod tidy

# Clean
clean:
	rm -rf bin/
	rm -rf tmp/
	go clean

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Setup project
setup: deps
	cp .env.example .env || echo "Create .env file manually"
	mkdir -p logs

# Docker build
docker-build:
	docker build -t imdb-performance-optimization .

# Docker run
docker-run:
	docker run -p 5000:5000 imdb-performance-optimization
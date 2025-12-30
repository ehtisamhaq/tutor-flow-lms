.PHONY: build run dev test clean docker-up docker-down migrate lint swagger

# Build the application
build:
	go build -o bin/server cmd/server/main.go

# Run the application
run: build
	./bin/server

# Run with hot reload (requires air)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Installing air..."; \
		go install github.com/air-verse/air@latest; \
		air; \
	fi

# Run tests
test:
	go test -v -cover ./...

# Run tests with coverage report
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker commands
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Database commands
migrate:
	go run cmd/server/main.go migrate

# Generate Swagger docs (requires swag)
swagger:
	@if command -v swag > /dev/null; then \
		swag init -g cmd/server/main.go -o docs; \
	else \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
		swag init -g cmd/server/main.go -o docs; \
	fi

# Lint code (requires golangci-lint)
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Download dependencies
deps:
	go mod download
	go mod tidy

# Create uploads directory
setup:
	mkdir -p uploads/videos uploads/images uploads/documents
	cp config/config.example.yaml config/config.yaml

# Help
help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Build and run the application"
	@echo "  make dev          - Run with hot reload"
	@echo "  make test         - Run tests"
	@echo "  make docker-up    - Start Docker containers"
	@echo "  make docker-down  - Stop Docker containers"
	@echo "  make swagger      - Generate Swagger documentation"
	@echo "  make lint         - Lint the code"
	@echo "  make setup        - Initial project setup"

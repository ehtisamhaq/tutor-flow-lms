# TutorFlow LMS - Server

A professional Learning Management System built with Go, Echo framework, and PostgreSQL.

## Features

- **Multi-role authentication**: Admin, Manager, Tutor, Student
- **Course management**: Create, publish, and manage courses
- **Content security**: DRM-protected video content with HLS encryption
- **E-commerce**: Shopping cart, Stripe payments, coupons
- **Assessment**: Quizzes and assignments with auto-grading
- **Learning paths**: Curated course sequences
- **Reviews & ratings**: Course feedback system
- **Notifications**: In-app and email notifications

## Tech Stack

- **Backend**: Go 1.22+ with Echo v4
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **ORM**: GORM v2
- **Auth**: JWT with refresh tokens
- **Payments**: Stripe

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make (optional)

### Setup

1. Clone the repository:

   ```bash
   cd tutorflow-server
   ```

2. Copy config file:

   ```bash
   cp config/config.example.yaml config/config.yaml
   ```

3. Start database services:

   ```bash
   make docker-up
   # or
   docker-compose up -d
   ```

4. Install dependencies:

   ```bash
   go mod download
   ```

5. Run the server:
   ```bash
   make run
   # or
   go run cmd/server/main.go
   ```

The API will be available at `http://localhost:8080`

### Development

For hot-reload during development:

```bash
make dev
```

### API Endpoints

| Endpoint                     | Description       |
| ---------------------------- | ----------------- |
| `GET /health`                | Health check      |
| `POST /api/v1/auth/register` | Register new user |
| `POST /api/v1/auth/login`    | Login             |
| `POST /api/v1/auth/refresh`  | Refresh token     |
| `GET /api/v1/auth/me`        | Get current user  |

### Makefile Commands

```bash
make build        # Build the application
make run          # Build and run
make dev          # Run with hot reload
make test         # Run tests
make docker-up    # Start Docker containers
make docker-down  # Stop Docker containers
make swagger      # Generate Swagger docs
make lint         # Lint the code
```

## Project Structure

```
tutorflow-server/
├── cmd/server/         # Application entrypoint
├── internal/
│   ├── domain/         # Business entities
│   ├── usecase/        # Business logic
│   ├── repository/     # Data access
│   ├── handler/        # HTTP handlers
│   ├── middleware/     # Custom middleware
│   └── pkg/            # Shared utilities
├── migrations/         # SQL migrations
├── config/             # Configuration files
├── docs/               # API documentation
└── scripts/            # Utility scripts
```

## License

MIT

# TutorFlow LMS - Server

A professional, high-performance Learning Management System built with Go, Echo framework, and PostgreSQL.

## 🚀 Features

- **Multi-role Authentication**: Secure access for Admin, Manager, Tutor, and Student roles with JWT & Refresh Tokens.
- **Course & Content Management**: Comprehensive system for creating courses, modules, and lessons with support for HLS encrypted video.
- **Learning Paths**: Structured curriculum sequences enabling students to follow curated educational journeys.
- **Assessment System**: Advanced quizzes and assignments with auto-grading, time limits, and feedback loops.
- **E-commerce Engine**: Integrated shopping cart, wishlist, Stripe payment gateway, and dynamic coupon/discount system.
- **Direct Messaging**: Real-time communication between students and instructors with file attachments.
- **Discussion Forums**: Course-specific Q&A and discussion threads to foster community learning.
- **Announcement System**: Course-wide updates with integrated Web Push and Email notifications.
- **Automated Certificates**: Dynamic PDF certificate generation upon course completion.
- **Advanced Search**: Full-text search across courses with filtering, facets, and autocomplete suggestions.
- **Reports & Analytics**: Detailed sales reports, student progress tracking, and engagement metrics for admins and instructors.

## 🛠 Tech Stack

- **Backend**: Go 1.25.3 with [Echo v4](https://echo.labstack.com/)
- **Database**: PostgreSQL 16
- **Persistence**: [GORM v2](https://gorm.io/)
- **Documentation**: [Swagger / OpenAPI 3.0](https://swagger.io/)
- **Cache**: Redis 7
- **Authentication**: JWT with Refresh Tokens
- **Payments**: Stripe
- **Storage**: Cloudinary / Local Storage
- **Logging**: Uber [Zap](https://github.com/uber-go/zap)

## ⚡️ Getting Started

### Prerequisites

- Go 1.25.3+
- Docker & Docker Compose
- Make (recommended)

### Installation & Setup

1. **Clone the repository**:

   ```bash
   git clone https://github.com/tutorflow/tutorflow-server.git
   cd tutorflow-server
   ```

2. **Run Initial Setup**:

   ```bash
   make setup
   ```

   _This will create necessary directories and copy the default configuration._

3. **Start Infrastructure**:

   ```bash
   make docker-up
   ```

   _Starts PostgreSQL and Redis in the background._

4. **Run the Application**:
   ```bash
   make dev
   ```
   _Runs the server with hot-reload enabled._

The API will be available at `http://localhost:8080`.

## 📖 API Documentation

The API is fully documented using Swagger annotations.

- **Interactive UI**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
- **JSON Spec**: `http://localhost:8080/swagger/doc.json`

To regenerate the documentation after making changes to handlers or models:

```bash
make swagger
```

## 📂 Project Structure

```
tutorflow-server/
├── cmd/                # Entry points for the application
│   └── server/         # Main server application
├── internal/           # Private application and library code
│   ├── domain/         # Business logic models (Entities)
│   ├── usecase/        # Business logic rules (Services)
│   ├── repository/     # Data access layer (Postgres/GORM)
│   ├── handler/        # HTTP controllers (Echo Handlers)
│   ├── middleware/     # Custom HTTP middleware
│   ├── service/        # External service integrations (Email, Stripe, Storage)
│   └── pkg/            # Reusable packages (JWT, Logger, Database)
├── config/             # Configuration files and loaders
├── docs/               # Generated Swagger documentation
├── migrations/         # Database migration files
└── uploads/            # Local storage for media assets
```

## 🧪 Development

### Running Tests

```bash
make test          # Run all tests
make test-coverage # Run tests and generate HTML coverage report
```

### Linting

```bash
make lint          # Run golangci-lint
```

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

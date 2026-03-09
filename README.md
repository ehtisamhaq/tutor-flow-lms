# tutor-flow-lms

A modern LMS built with Go (Backend) and Next.js (Frontend).

## Local Development (Hybrid Workflow)

To keep your machine cool and performance high, we run infrastructure in Docker and the application natively.

### 1. Start Infrastructure (Docker)

Run the databases and storage:

```bash
docker compose up -d
```

### 2. Seed Databases (Optional)

```bash
docker exec -i tutorflow_postgres psql -U postgres -d tutorflow < tutorflow-server/scripts/seed_users.sql
```

### 3. Start Backend (Go)

Navigate to the server directory and run the api:

```bash
cd tutorflow-server
go run cmd/server/main.go
```

### 4. Start Frontend (Next.js)

Navigate to the client directory and run the dev server:

```bash
cd tutorflow-client
bun dev
```

## Stripe Webhook (Optional)

To test payments locally:

```bash
stripe listen --forward-to localhost:8080/webhook
```

# tutor-flow-lms

docker compose up -d
docker compose up --build
docker exec -i tutorflow_postgres psql -U postgres -d tutorflow < tutorflow-server/scripts/seed_users.sql

<!-- stripe webhook  -->

stripe listen --forward-to localhost:8080/webhook

# Backend

cd tutorflow-server
docker compose up -d
make run

# Frontend

cd tutorflow-client
bun dev

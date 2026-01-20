# tutor-flow-lms

docker compose up --build

# Backend

cd tutorflow-server
docker compose up -d
make run

# Frontend

cd tutorflow-client
bun dev

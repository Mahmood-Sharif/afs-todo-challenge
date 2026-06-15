# Modular ToDo Application with Authentication

AFS x Reboot technical challenge project. This repository will become a modular,
multi-tier ToDo application with a React frontend, Go backend, and PostgreSQL
database.

## Tech Stack

- Frontend: React + Vite
- Backend: Go
- Database: PostgreSQL
- Orchestration: Docker Compose

## Architecture Overview

The project is split into three independent services:

- `frontend`: React + Vite app served from its own container on port `5173`.
- `backend`: Go REST API served from its own container on port `8080`.
- `postgres`: PostgreSQL database in its own container on the internal Docker
  network.

Docker Compose creates a shared network so services can communicate by service
name. PostgreSQL data is persisted with a named Docker volume.

The backend connects to PostgreSQL using the Compose service name `postgres` as
the database host. PostgreSQL is not exposed to the host machine in this phase;
only the backend can reach it through the Docker network.

## Backend Environment Variables

The backend reads configuration from environment variables:

- `PORT`: backend HTTP port
- `DB_HOST`: PostgreSQL host, set to `postgres` in Docker Compose
- `DB_PORT`: PostgreSQL port inside the Docker network
- `DB_USER`: PostgreSQL username
- `DB_PASSWORD`: PostgreSQL password
- `DB_NAME`: PostgreSQL database name
- `DB_SSLMODE`: PostgreSQL SSL mode

## Run with Docker Compose

From the project root:

```bash
copy .env.example .env
docker compose up --build
```

The services will be available at:

- Frontend: <http://localhost:5173>
- Backend health endpoint: <http://localhost:8080/api/health>

## Test the Health Endpoints

In another terminal, test the backend service health:

```bash
curl http://localhost:8080/api/health
```

Expected response:

```json
{
  "status": "ok",
  "service": "backend"
}
```

Test the backend-to-database connection:

```bash
curl http://localhost:8080/api/db-health
```

Expected response:

```json
{
  "status": "ok",
  "service": "database",
  "database": "postgres"
}
```

## Current Phase Status

Phase 2 is complete:

- Created project skeleton.
- Added Docker Compose orchestration for frontend, backend, and PostgreSQL.
- Added a minimal React + Vite frontend.
- Added a minimal Go backend with `GET /api/health`.
- Added persistent PostgreSQL storage through a Docker volume.
- Added backend configuration loading from environment variables.
- Added backend PostgreSQL connection and startup ping.
- Added `GET /api/db-health` for database connectivity checks.

Not included yet:

- Authentication
- Todo CRUD
- Database schema and migrations
- Frontend login, registration, or todo dashboard

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
- `postgres`: PostgreSQL database in its own container on port `5432`.

Docker Compose creates a shared network so services can communicate by service
name. PostgreSQL data is persisted with a named Docker volume.

## Run with Docker Compose

From the project root:

```bash
copy .env.example .env
docker compose up --build
```

The services will be available at:

- Frontend: <http://localhost:5173>
- Backend health endpoint: <http://localhost:8080/api/health>
- PostgreSQL: `localhost:5432`

## Test the Backend Health Endpoint

In another terminal, run:

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

## Current Phase Status

Phase 1 is complete:

- Created project skeleton.
- Added Docker Compose orchestration for frontend, backend, and PostgreSQL.
- Added a minimal React + Vite frontend.
- Added a minimal Go backend with `GET /api/health`.
- Added persistent PostgreSQL storage through a Docker volume.

Not included yet:

- Authentication
- Todo CRUD
- Database schema and migrations
- Frontend login, registration, or todo dashboard

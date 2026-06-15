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

## Database Schema

Phase 3 adds the initial PostgreSQL schema:

- `users`: stores application users with unique email addresses.
- `todos`: stores todo items owned by users.

The relationship is one user to many todos. Each todo row has a `user_id`
foreign key referencing `users(id)`. The foreign key uses `ON DELETE CASCADE`,
so deleting a user also deletes that user's todos.

Key integrity rules:

- `users.email` is unique.
- `users.name`, `users.email`, and `users.password_hash` are required.
- `todos.user_id` and `todos.title` are required.
- `todos.is_completed` defaults to `false`.
- `created_at` and `updated_at` default to the current timestamp.

## Migrations

SQL migrations live in `backend/migrations/`. The backend embeds these files and
runs pending migrations automatically on startup after connecting to PostgreSQL.

Applied migrations are tracked in the `schema_migrations` table. This keeps
startup idempotent: restarting the backend does not recreate or reapply already
recorded migrations.

## Backend Environment Variables

The backend reads configuration from environment variables:

- `PORT`: backend HTTP port
- `DB_HOST`: PostgreSQL host, set to `postgres` in Docker Compose
- `DB_PORT`: PostgreSQL port inside the Docker network
- `DB_USER`: PostgreSQL username
- `DB_PASSWORD`: PostgreSQL password
- `DB_NAME`: PostgreSQL database name
- `DB_SSLMODE`: PostgreSQL SSL mode
- `JWT_SECRET`: secret used to sign authentication JWTs
- `JWT_EXPIRY_HOURS`: number of hours before auth cookies expire

## Authentication

Phase 4 adds backend authentication with HTTP-only cookie-based JWTs.

Passwords are hashed with bcrypt before storage. This means plain-text
passwords are never stored in PostgreSQL, and `password_hash` is never returned
from the API.

Login creates a signed JWT and stores it in an HTTP-only cookie named
`afs_auth_token`. HTTP-only cookies help keep tokens out of browser JavaScript,
which reduces exposure if frontend code is ever compromised. The cookie uses
`SameSite=Lax`, which is suitable for local development and same-site app flows.

Auth endpoints:

- `POST /api/auth/register`: create a user
- `POST /api/auth/login`: authenticate and set the auth cookie
- `POST /api/auth/logout`: clear the auth cookie
- `GET /api/auth/me`: return the authenticated user

The `/api/auth/me` route is protected by middleware that validates the auth
cookie and places the authenticated `user_id` into the request context.

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

Test that the required schema exists:

```bash
curl http://localhost:8080/api/schema-health
```

Expected response:

```json
{
  "status": "ok",
  "service": "schema",
  "tables": ["users", "todos"]
}
```

## Test Authentication

Register a user:

```bash
curl -i -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Mahmood\",\"email\":\"mahmood@example.com\",\"password\":\"password123\"}"
```

Log in and save the HTTP-only cookie:

```bash
curl -i -c cookies.txt -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"mahmood@example.com\",\"password\":\"password123\"}"
```

Get the current authenticated user:

```bash
curl -i -b cookies.txt http://localhost:8080/api/auth/me
```

Log out:

```bash
curl -i -b cookies.txt -c cookies.txt -X POST http://localhost:8080/api/auth/logout
```

Confirm the session was cleared:

```bash
curl -i -b cookies.txt http://localhost:8080/api/auth/me
```

Expected result after logout: HTTP `401` with `{"error":"unauthorized"}`.

## Reset the Database Volume

To remove the local PostgreSQL volume and force migrations to run against a
fresh database:

```bash
docker compose down -v
docker compose up --build
```

## Current Phase Status

Phase 4 is complete:

- Created project skeleton.
- Added Docker Compose orchestration for frontend, backend, and PostgreSQL.
- Added a minimal React + Vite frontend.
- Added a minimal Go backend with `GET /api/health`.
- Added persistent PostgreSQL storage through a Docker volume.
- Added backend configuration loading from environment variables.
- Added backend PostgreSQL connection and startup ping.
- Added `GET /api/db-health` for database connectivity checks.
- Added SQL migrations for `users` and `todos`.
- Added automatic migration execution on backend startup.
- Added `GET /api/schema-health` for schema verification.
- Added backend registration, login, logout, and current-user auth endpoints.
- Added bcrypt password hashing.
- Added HTTP-only cookie-based JWT authentication.
- Added auth middleware for protected backend routes.

Not included yet:

- Todo CRUD
- Frontend login, registration, or todo dashboard

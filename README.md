# SkillMatch

An AI-powered job-search assistant with **persistent agentic memory**. Built
for the CockroachDB × AWS hackathon and migrated to PostgreSQL: PostgreSQL
(pgvector) is the memory layer, Google Gemini is the AI, and resume files are
stored on the server's local filesystem.

## Stack

| Layer       | Technology                        |
| ----------- | --------------------------------- |
| Frontend    | React + Vite + TypeScript         |
| Backend     | Go (`net/http`)                   |
| Database    | PostgreSQL (pgvector)             |
| Object store| Local filesystem (served at `/storage`) |
| AI          | Google Gemini (generateContent REST) |
| Auth        | JWT (HS256)                       |

## Repository layout

```
backend/   Go API server (handlers, services, repositories, clients, migrations)
frontend/  React + Vite application
docs/      architecture, API, database, contributor docs
```

## Prerequisites

- Go 1.24+
- Node.js 18+ and npm
- A PostgreSQL 16+ database with the `pgvector` extension (see
  [`scripts/setup_postgres.sh`](scripts/setup_postgres.sh) for a one-command
  local container)
- A Google Gemini API key (`GEMINI_API_KEY`)

## Setup

### 1. Database

The server applies migrations automatically on startup (see
[`backend/migrations/runner.go`](backend/migrations/runner.go)); applied
versions are tracked in the `schema_migrations` table. The `pgvector`
extension is enabled by migration `003`.

For local development, spin up a pgvector-enabled container:

```sh
./scripts/setup_postgres.sh
```

or create the database manually:

```sql
CREATE DATABASE skillmatch;
```

### 2. Backend

```sh
cd backend
cp .env.example .env   # then fill in the values below
go run ./cmd/api
```

Environment variables (`backend/.env`):

| Variable          | Required | Default            | Description                                    |
| ----------------- | -------- | ------------------ | ---------------------------------------------- |
| `PORT`            | no       | `8080`             | HTTP listen port                               |
| `DATABASE_URL`    | yes      | —                  | PostgreSQL connection string (postgres://)     |
| `JWT_SECRET`      | yes      | —                  | JWT signing secret                             |
| `CORS_ALLOWED_ORIGIN` | no   | `http://localhost:5173` | allowed browser origin(s)                 |
| `STORAGE_DIR`     | no       | `./data/resumes`   | directory where resume files are stored        |
| `GEMINI_API_KEY`  | no*      | —                  | Google Gemini API key (*required for chat/tailor) |
| `GEMINI_MODEL`    | no       | `gemini-2.5-flash` | Gemini model used for chat and CV tailoring    |

> Note: chat (`/api/chat`) and CV tailoring (`/api/tailor`) endpoints are only
> registered when `GEMINI_API_KEY` is set.

### 3. Frontend

```sh
cd frontend
npm install
npm run dev
```

The frontend reads `VITE_API_BASE_URL` (default `http://localhost:8080/api`);
override it if the backend runs elsewhere (e.g. `http://localhost:8090/api`).

## Running

- Backend: `go run ./cmd/api` — listens on `PORT`, applies migrations, serves
  `/health` (pings PostgreSQL and local storage; 503 when degraded).
- Frontend: `npm run dev` inside `frontend/`.
- Root scripts: `npm run dev` / `npm run build` (build the frontend).

## Testing

```sh
cd backend
go test ./...                # unit tests (no infrastructure required)

# integration tests against a live database:
export TEST_DATABASE_URL='postgres://...'
go test -tags integration ./...
```

## API endpoints

See [docs/API.md](docs/API.md). Implemented today: health, auth
(register/login), resume management, job search, recommendations, saved jobs,
applications, chat, and CV tailoring.

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Data model](docs/DATABASE.md)
- [API reference](docs/API.md)
- [Contributors](docs/CONTRIBUTORS.md)

## Known limitations

- The chat assistant keeps a single per-user memory stream on the server; the
  frontend's multiple named conversations are stored locally (localStorage) and
  are not yet modelled as separate threads on the backend.
- Resume parsing supports `.pdf`, `.docx`, and `.txt`. Legacy `.doc` files are
  rejected with a clear message (convert to PDF/DOCX/TXT).
- When `JWT_SECRET` is unset the backend refuses to start; set a stable secret
  so tokens survive restarts.
- Resume upload is capped at 5 MB and accepts `.pdf`, `.doc`, `.docx`, `.txt`.

## License

MIT — see [LICENSE](LICENSE).

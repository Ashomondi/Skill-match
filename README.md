# SkillMatch

An AI-powered job-search assistant with **persistent agentic memory**. Built
for the CockroachDB × AWS hackathon and migrated to PostgreSQL: PostgreSQL
(pgvector) is the memory layer, Amazon Bedrock is the AI, and resume files
are stored in S3 (MinIO locally).

## Stack

| Layer       | Technology                        |
| ----------- | --------------------------------- |
| Frontend    | React + Vite + TypeScript         |
| Backend     | Go (`net/http`)                   |
| Database    | PostgreSQL (pgvector)             |
| Object store| Amazon S3 / MinIO (S3-compatible) |
| AI          | Amazon Bedrock (planned)          |
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
- S3-compatible object storage (MinIO for local dev, or Amazon S3)

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

| Variable             | Required | Default     | Description                                  |
| -------------------- | -------- | ----------- | -------------------------------------------- |
| `PORT`               | no       | `8080`      | HTTP listen port                             |
| `DATABASE_URL`       | yes      | —           | PostgreSQL connection string (postgres://)  |
| `JWT_SECRET`         | no*      | ephemeral   | JWT signing secret (*set for stable tokens)  |
| `JWT_EXPIRATION`     | no       | `24h`       | token lifetime (`time.ParseDuration` format) |
| `AWS_REGION`         | no       | `us-east-1` | region used for signing                      |
| `S3_BUCKET_NAME`     | yes**    | —           | bucket for resume files (**required for resume API) |
| `S3_ENDPOINT`        | no       | —           | custom endpoint for MinIO/LocalStack; empty = real AWS S3 |
| `AWS_ACCESS_KEY_ID`  | no       | —           | static credentials (used when `S3_ENDPOINT` is set) |
| `AWS_SECRET_ACCESS_KEY` | no    | —           |                                            |
| `S3_FORCE_PATH_STYLE`| no       | `true`      | path-style addressing (required for MinIO)   |

### 3. Object store (MinIO for local dev)

```sh
docker run -d -p 9000:9000 -p 9001:9001 \
  -e "MINIO_ROOT_USER=minioadmin" -e "MINIO_ROOT_PASSWORD=minioadmin123" \
  quay.io/minio/minio server /data --console-address ":9001"
```

Create a bucket (e.g. `initone`) in the console at `http://localhost:9001`, then
point the backend at it:

```
S3_ENDPOINT=http://localhost:9000
S3_BUCKET_NAME=initone
AWS_ACCESS_KEY_ID=minioadmin
AWS_SECRET_ACCESS_KEY=minioadmin123
```

For real AWS S3, leave `S3_ENDPOINT` empty and rely on the default credential
chain (env vars, `~/.aws/credentials`, or an IAM role).

### 4. Frontend

```sh
cd frontend
npm install
npm run dev
```

The frontend reads `VITE_API_BASE_URL` (default `http://localhost:8080/api`);
override it if the backend runs elsewhere (e.g. `http://localhost:8090/api`).
Optional demo-login vars: `VITE_DEMO_AUTH_ENABLED`, `VITE_DEMO_EMAIL`,
`VITE_DEMO_PASSWORD`, `VITE_DEMO_NAME`.

## Running

- Backend: `go run ./cmd/api` — listens on `PORT`, applies migrations, serves
  `/health` (pings PostgreSQL; 503 when degraded).
- Frontend: `npm run dev` inside `frontend/`.
- Root scripts: `npm run dev` / `npm run build` (build the frontend).

## Testing

```sh
cd backend
go test ./...                # unit tests (no infrastructure required)

# integration tests against live infra:
export TEST_DATABASE_URL='postgres://...'
export TEST_S3_ENDPOINT='http://localhost:9000'
export TEST_S3_BUCKET='initone'
export TEST_S3_ACCESS_KEY='minioadmin'
export TEST_S3_SECRET_KEY='minioadmin123'
go test -tags integration ./...
```

## API endpoints

See [docs/API.md](docs/API.md). Implemented today: health, auth
(register/login), and resume management. Chat, job search, recommendations,
saved jobs, and applications endpoints are planned.

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Data model](docs/DATABASE.md)
- [API reference](docs/API.md)
- [Contributors](docs/CONTRIBUTORS.md)

## Known limitations

- The chat, job-search/recommendation, saved-jobs, and application-tracking API
  endpoints are not yet implemented; the frontend pages for those features are
  partially backed by mock/hardcoded data.
- Amazon Bedrock integration (chat + embeddings) is pending; the memory layer
  is fully implemented and tested against PostgreSQL.
- When `JWT_SECRET` is unset, an ephemeral secret is generated per boot, so
  existing tokens are invalidated on restart.
- Resume upload is capped at 5 MB and accepts `.pdf`, `.doc`, `.docx`, `.txt`.

## License

MIT — see [LICENSE](LICENSE).

# SkillMatch Architecture

## 1. Overview

SkillMatch is an AI-powered job-search assistant that uses PostgreSQL (with
the pgvector extension) as a **persistent memory layer**. Users register,
upload a resume, chat with an AI assistant that remembers prior conversations,
and get personalized job recommendations derived from resume/job embeddings
and interaction history.

The core design idea: the AI (Bedrock) is **stateless**, and PostgreSQL is
where memory actually lives — conversation turns, vector embeddings, and
structured user data. Deleting a user cascades across all of it.

## 2. High-level flow

1. User signs up / logs in → JWT.
2. User uploads a resume → file bytes to S3, metadata row to PostgreSQL.
3. Backend extracts text and stores an embedding (Titan, 1536 dims).
4. User chats with the assistant → past turns + similar memories are loaded
   from PostgreSQL, packaged into the prompt, sent to Bedrock, and the reply
   is stored back as a new turn.
5. Jobs are embedded too; the matching engine compares resume/query vectors
   (pgvector HNSW index) plus interaction history.
6. User saves jobs and tracks applications; the dashboard aggregates it all.

## 3. Layer diagram

```
React (Vite) frontend
        │  REST/JSON + JWT (Authorization: Bearer)
        ▼
net/http server  ── middleware: Recovery → Logging → CORS
        │
        ├─ handlers   (auth, resume, …)   — HTTP concerns, user id from context
        │        │
        │        ▼
        │  services  (auth, resume, job interactions, applications, …)
        │        │       business rules + ownership checks
        │        ▼
        │  repositories  (user, resume, conversation, embedding, …) — SQL only
        │        │
        │        └────────────►  PostgreSQL (pgx pool)
        │
        ├─ clients  (S3/MinIO, PostgreSQL pool, Bedrock planned)
        │        └────────────►  S3 / MinIO
        │
        └─ migrations  (embedded SQL, auto-applied at startup)
```

### Dependencies

- `handlers` → `services` → `repositories` → PostgreSQL.
- `handlers`/`services` depend on `repositories` **interfaces** defined in the
  service layer, so services and handlers are unit-testable with fakes.
- `clients` wraps external SDKs (pgx pool, AWS S3). Repositories never touch
  the pool directly via callers; they're constructed with the shared pool.

## 4. Key decisions

| Decision | Rationale |
| -------- | --------- |
| `net/http` + `http.ServeMux` (Go 1.22 method/pattern routing) | stdlib, zero deps |
| `models` package as the canonical domain types | repositories no longer define their own divergent structs |
| Auth middleware validates JWT and injects `user_id`/`email` into context | handlers/services stay context-based, no global state |
| S3 client supports custom endpoint + static creds + path style | MinIO/LocalStack are S3-compatible and used for dev |
| Migrations run at startup via an embedded runner + `schema_migrations` | no manual migration step; idempotent + versioned |
| Migrations run as multi-statement exec, not in a pgx transaction | files are written idempotently and are safe to re-run after a partial failure |
| Embeddings are polymorphic (`source_type` + `source_id`) | one table serves resume, conversation, and job vectors |
| Vector ANN via `ORDER BY vector <=> $1 LIMIT n` | backed by a pgvector HNSW index (`vector_cosine_ops`) |

## 5. Authentication & authorization

- Passwords hashed with bcrypt (`golang.org/x/crypto`).
- JWT (HS256) issued on register/login; `{ token, user }` in both responses.
- `middleware.Auth` wraps protected routes; `UserIDFromContext` gives handlers
  the authenticated user.
- **Ownership**: every service that fetches/deletes user data re-checks that
  the resource's `user_id` matches the authenticated user (resumes,
  interactions, applications) — enforced in the service layer, not just the SQL.

## 6. Storage

- **PostgreSQL** — users, resumes (metadata), conversations, embeddings,
  jobs, saved jobs, interactions, applications, schema_migrations. See
  [DATABASE.md](DATABASE.md).
- **S3/MinIO** — resume file bytes at `resumes/{userID}/{fileID}`, accessed via
  short-lived presigned URLs.

## 7. Testing strategy

- **Unit tests** (no infra): services against fake repositories/storage,
  handlers via `httptest` behind the real auth middleware, middleware JWT
  flows, clients config validation. `go test ./...`.
- **Integration tests** (`//go:build integration`, env-guarded): real
  PostgreSQL + MinIO; cover memory workflow, resume storage round-trip, and
  interaction/application persistence. Run with `TEST_DATABASE_URL` and
  `TEST_S3_*`. `go test -tags integration ./...`.

## 8. Deployment notes

- Build the backend binary, set `DATABASE_URL`/`S3_*`/`JWT_SECRET`, run it —
  migrations apply automatically. See [README.md](../README.md).
- The frontend is a static Vite build that talks to the backend over HTTP;
  CORS is open for development.

# Contributors

SkillMatch is built for the CockroachDB × AWS hackathon by a five-person team.
Backend work is split into the layers below; every layer uses the canonical
`models` package as its source of truth.

## Team

| Name     | Focus area (milestone issues) |
| -------- | ----------------------------- |
| Sospeter | Database/repository layer; application tracking; job interactions; testing |
| Ashley   | Authentication, resume API, chat/job services |
| Evans    | Backend foundation, S3, Bedrock/MCP clients, matching engine, deployment |
| Ian      | Frontend pages and components |
| Emma     | Frontend services, hooks, and integration |

## How to contribute

1. **Follow the architecture.** Handlers → services → repositories. Handlers
   read the authenticated user from `middleware.UserIDFromContext`; services
   enforce ownership; repositories are the only layer issuing SQL. See
   [ARCHITECTURE.md](ARCHITECTURE.md).
2. **Canonical types live in `models`.** Add/update the model first, then the
   repository, then the service.
3. **Schema changes are migrations.** Add a new numbered file in
   `backend/migrations/` (e.g. `007_*.sql`). The runner applies it on startup;
   keep statements idempotent (`IF NOT EXISTS`) where possible.
4. **Tests.** Unit tests with fakes for services/handlers (`go test ./...`).
   For anything touching the DB or S3, add a build-tagged integration test and
   run it with `TEST_DATABASE_URL`/`TEST_S3_*` set
   (`go test -tags integration ./...`).
5. **Formatting.** `gofmt` and `go vet ./...` must be clean; the frontend must
   pass `tsc --noEmit`.

## Getting started

See [README.md](../README.md) for setup and the [data model](DATABASE.md) for
the schema.
